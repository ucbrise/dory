import re
import matplotlib.pyplot as plt
import sys
import numpy as np
from matplotlib.ticker import FuncFormatter
import math
from collections import defaultdict
from matplotlib.patches import Patch
from matplotlib.ticker import ScalarFormatter
import brewer2mpl

bmap1 = brewer2mpl.get_map('Set1', 'Qualitative', 7)
bmap2 = brewer2mpl.get_map('Dark2', 'Qualitative', 7)
hash_colors = bmap1.mpl_colors
mix_colors = bmap2.mpl_colors

dory_1_color = hash_colors[2]
dory_2_color = hash_colors[4]
dory_4_color = hash_colors[1]
oram_color =  hash_colors[0]

dory_1_y = []
with open("out/dory_throughput_1_1_9.dat", 'r') as f:
    for i, line in enumerate(f):
        dory_1_y.append(float(line))

# N = 10000, n = 100
print(dory_1_y)

dory_2_y = []
with open("out/dory_throughput_2_1_9.dat", 'r') as f:
    for i, line in enumerate(f):
        dory_2_y.append(float(line))

# N = 10000, n = 100
print(dory_2_y)

dory_4_y = []
with open("out/dory_throughput_4_1_9.dat", 'r') as f:
    for i, line in enumerate(f):
        dory_4_y.append(float(line))

oram_y = []
with open("ref/oram_throughput_1_9.dat", 'r') as f:
    for i, line in enumerate(f):
        oram_y.append(float(line))

# N = 10000, n = 100

x = [2**i for i in range(10,21)]
oram_x = [2**i for i in range(10, len(oram_y) + 10)]

fig = plt.figure(figsize = (8,8))
ax = fig.add_subplot(111)
ax.loglog(x, dory_1_y, label=r"DORY ($p=1$)", basex=2, basey=10, color=dory_1_color, linewidth=1)
ax.loglog(x, dory_2_y, label=r"DORY ($p=2$)", basex=2, basey=10, color=dory_2_color, linewidth=1)
ax.loglog(x, dory_4_y, label=r"DORY ($p=4$)", basex=2, basey=10, color=dory_4_color, linewidth=1)
ax.loglog(oram_x, oram_y, label=r"PathORAM baseline", basex=2, basey=10, color=oram_color, linewidth=1, linestyle="dashed")
ax.set_xlabel("\# Documents")
ax.set_ylabel("Operations/sec")
ax.set_xticks([1024, 32768, 1048576])
ax.set_yticks([100, 1, 0.01, 0.0001])

plt.legend()
plt.tight_layout()
plt.savefig("out/fig10a.png")
plt.show()
