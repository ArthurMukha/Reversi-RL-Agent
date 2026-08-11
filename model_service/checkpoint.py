"""Загрузка весов модели с диска.

Слой знает про файлы и про torch, но ничего не знает ни про gRPC, ни про
правила игры. Наверх отдаёт готовую к инференсу модель и её метаданные.

Формат файла:
    {"state_dict": ..., "arch": {...}, "meta": {...}}

Проверить руками:
    python -m model_service.checkpoint
"""

from dataclasses import dataclass
from pathlib import Path

import torch

from model_service.net import PolicyNetwork

CHECKPOINTS_DIR = Path(__file__).parent / "checkpoints"


@dataclass
class LoadedModel:
    model_id: str
    model: PolicyNetwork  # уже в eval()
    arch: dict
    meta: dict


def derive_arch(state_dict) -> dict:
    """Выводит архитектуру из форм тензоров."""
    channels = state_dict["stem.0.weight"].shape[0]

    blocks = {
        int(key.split(".")[1]) for key in state_dict if key.startswith("res_blocks.")
    }

    return {"channels": channels, "n_res_blocks": len(blocks)}


def load_checkpoint(path: Path) -> LoadedModel:
    ckpt = torch.load(path, map_location="cpu", weights_only=True)

    if "state_dict" not in ckpt:
        raise ValueError(
            f"{path.name}: голый state_dict без arch/meta — файл нужно обернуть"
        )

    state_dict, arch, meta = ckpt["state_dict"], ckpt["arch"], ckpt["meta"]

    derived = derive_arch(state_dict)
    if derived != arch:
        raise ValueError(f"{path.name}: в файле записано {arch}, а по весам {derived}")

    model = PolicyNetwork(**arch)
    model.load_state_dict(state_dict, strict=True)

    # без eval() BatchNorm нормализует по единственной позиции в батче,
    # и обученная модель играет как случайная — при этом всё «работает»
    model.eval()

    return LoadedModel(
        model_id=meta.get("model_id") or path.stem,
        model=model,
        arch=arch,
        meta=meta,
    )


def _signature(model: PolicyNetwork) -> float:
    """Грубый отпечаток весов — чтобы заметить два одинаковых файла."""
    with torch.inference_mode():
        return float(sum(p.sum() for p in model.parameters()))


def _smoke_test(loaded: LoadedModel) -> None:
    with torch.inference_mode():
        policy, value = loaded.model(torch.zeros(1, 2, 8, 8))

    assert policy.shape == (1, 64), f"policy: {tuple(policy.shape)}, ждали (1, 64)"
    assert value.shape == (1, 1), f"value: {tuple(value.shape)}, ждали (1, 1)"

    v = value.item()
    assert -1.0 <= v <= 1.0, f"value вне [-1, 1]: {v}"

    print(f"  model_id  : {loaded.model_id}")
    print(
        f"  arch      : {loaded.arch}, отпечаток весов {_signature(loaded.model):+.4f}"
    )
    print(f"  forward   : policy {tuple(policy.shape)}, value {v:+.4f}")
    print(f"  meta      : {loaded.meta}")


if __name__ == "__main__":
    files = sorted(CHECKPOINTS_DIR.glob("*.pt"))
    if not files:
        raise SystemExit(f"в {CHECKPOINTS_DIR} нет ни одного .pt")

    ok: list[LoadedModel] = []

    for path in files:
        print(f"\n{path.name}")
        try:
            loaded = load_checkpoint(path)
        except (ValueError, KeyError) as err:
            print(f"  ПРОПУСК   : {err}")
            continue

        _smoke_test(loaded)
        ok.append(loaded)

    print(f"\nзагружено моделей: {len(ok)} из {len(files)}")

    signatures = {m.model_id: _signature(m.model) for m in ok}
    if len(set(signatures.values())) != len(signatures):
        print(f"ВНИМАНИЕ: веса совпадают у разных моделей — {signatures}")
    else:
        print("веса всех моделей различаются")
