import torch
import torch.nn.functional as F
from torch import nn


class PolicyNetwork(nn.Module):
    def __init__(self, channels=64, n_res_blocks=2):
        super().__init__()

        self.arch = {
            "channels": channels,
            "n_res_blocks": n_res_blocks,
        }

        self.stem = nn.Sequential(
            nn.Conv2d(2, channels, kernel_size=3, padding=1, bias=False),
            nn.BatchNorm2d(channels),
            nn.ReLU(),
        )
        self.res_blocks = nn.ModuleList(
            [ResBlock(channels) for _ in range(n_res_blocks)]
        )

        # policy head
        self.policy_conv = nn.Conv2d(channels, 2, kernel_size=1, bias=False)
        self.policy_bn = nn.BatchNorm2d(2)
        self.policy_fc = nn.Linear(2 * 8 * 8, 64)

        # value head
        self.value_conv = nn.Conv2d(channels, 1, kernel_size=1, bias=False)
        self.value_bn = nn.BatchNorm2d(1)
        self.value_fc1 = nn.Linear(8 * 8, 64)
        self.value_fc2 = nn.Linear(64, 1)

    def forward(self, x):
        x = self.stem(x)
        for block in self.res_blocks:
            x = block(x)

        policy = F.relu(self.policy_bn(self.policy_conv(x)))
        policy = self.policy_fc(policy.flatten(1))

        value = F.relu(self.value_bn(self.value_conv(x)))
        value = F.relu(self.value_fc1(value.flatten(1)))
        value = torch.tanh(self.value_fc2(value))

        return policy, value


class ResBlock(nn.Module):
    def __init__(self, channels):
        super().__init__()
        self.conv1 = nn.Conv2d(channels, channels, 3, padding=1, bias=False)
        self.bn1 = nn.BatchNorm2d(channels)
        self.conv2 = nn.Conv2d(channels, channels, 3, padding=1, bias=False)
        self.bn2 = nn.BatchNorm2d(channels)

    def forward(self, x):
        residual = x
        out = F.relu(self.bn1(self.conv1(x)))
        out = self.bn2(self.conv2(out))
        return F.relu(out + residual)
