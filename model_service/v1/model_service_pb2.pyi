from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Cell(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    CELL_EMPTY: _ClassVar[Cell]
    CELL_WHITE: _ClassVar[Cell]
    CELL_BLACK: _ClassVar[Cell]
CELL_EMPTY: Cell
CELL_WHITE: Cell
CELL_BLACK: Cell

class State(_message.Message):
    __slots__ = ("board", "current", "legal_moves")
    BOARD_FIELD_NUMBER: _ClassVar[int]
    CURRENT_FIELD_NUMBER: _ClassVar[int]
    LEGAL_MOVES_FIELD_NUMBER: _ClassVar[int]
    board: _containers.RepeatedScalarFieldContainer[Cell]
    current: Cell
    legal_moves: _containers.RepeatedCompositeFieldContainer[Move]
    def __init__(self, board: _Optional[_Iterable[_Union[Cell, str]]] = ..., current: _Optional[_Union[Cell, str]] = ..., legal_moves: _Optional[_Iterable[_Union[Move, _Mapping]]] = ...) -> None: ...

class SelectMoveRequest(_message.Message):
    __slots__ = ("state", "model_id", "temperature")
    STATE_FIELD_NUMBER: _ClassVar[int]
    MODEL_ID_FIELD_NUMBER: _ClassVar[int]
    TEMPERATURE_FIELD_NUMBER: _ClassVar[int]
    state: State
    model_id: str
    temperature: float
    def __init__(self, state: _Optional[_Union[State, _Mapping]] = ..., model_id: _Optional[str] = ..., temperature: _Optional[float] = ...) -> None: ...

class Move(_message.Message):
    __slots__ = ("row", "col")
    ROW_FIELD_NUMBER: _ClassVar[int]
    COL_FIELD_NUMBER: _ClassVar[int]
    row: int
    col: int
    def __init__(self, row: _Optional[int] = ..., col: _Optional[int] = ...) -> None: ...

class SelectMoveResponse(_message.Message):
    __slots__ = ("move", "value", "policy", "model_id")
    MOVE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    POLICY_FIELD_NUMBER: _ClassVar[int]
    MODEL_ID_FIELD_NUMBER: _ClassVar[int]
    move: Move
    value: float
    policy: _containers.RepeatedScalarFieldContainer[float]
    model_id: str
    def __init__(self, move: _Optional[_Union[Move, _Mapping]] = ..., value: _Optional[float] = ..., policy: _Optional[_Iterable[float]] = ..., model_id: _Optional[str] = ...) -> None: ...
