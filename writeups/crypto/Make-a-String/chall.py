from functools import reduce
from operator import xor
from collections import deque

class LFSR():
    def __init__(self, state):
        self.state = deque(state)
        self.taps = [63, 61, 60, 0]
        for i in range(2025):
            self.step()

    def step(self):
        res = reduce(xor, [self.state[i] for i in self.taps])
        self.state.popleft()
        self.state.append(res)
        return res

    def next_byte(self):
        b = 0
        for _ in range(8):
            b = (b<<1) | self.step()
        return b

state = input("Give me your state (64-bit bitstring): ")

if not len(state) == 64 and set(state).issubset(set("01")):
    print("Bad format, provide the state as a 64-bit bitstring!")
    exit()

state = list(map(int, state))
lfsr = LFSR(state)

if b"EPTCTF25" == bytes(lfsr.next_byte() for _ in range(8)):
    print("Good job!")
    print(open("flag.txt").read())
else:
    print("Too bad")



