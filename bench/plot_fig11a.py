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
with open("ref/dory_throughput_1_1_9.dat", 'r') as f:
    for i, line in enumerate(f):
        dory_1_y.append(float(line))

# N = 10000, n = 100
print(dory_1_y)

dory_2_y = []
with open("ref/dory_throughput_2_1_9.dat", 'r') as f:
    for i, line in enumerate(f):
        dory_2_y.append(float(line))

# N = 10000, n = 100
print(dory_2_y)

dory_4_y = []
with open("ref/dory_throughput_4_1_9.dat", 'r') as f:
    for i, line in enumerate(f):
        dory_4_y.append(float(line))

#

# N = 10000, n = 100

x = [2**i for i in range(10,21)]

plot_1_y = [1] * len(dory_1_y)
print(plot_1_y)
plot_2_y = [dory_2_y[i] / dory_1_y[i] for i in range(len(dory_1_y))]
print(plot_2_y)
plot_4_y = [dory_4_y[i] / dory_1_y[i] for i in range(len(dory_1_y))]
print(plot_4_y)


fig = plt.figure(figsize = (8,8))
ax = fig.add_subplot(111)
print(x)
ax.plot(x, plot_1_y, label=r"DORY ($p=1$)", color=dory_1_color, linewidth=1)
ax.plot(x, plot_2_y, label=r"DORY ($p=2$)", color=dory_2_color, linewidth=1)
ax.plot(x, plot_4_y, label=r"DORY ($p=4$)", color=dory_4_color, linewidth=1)
ax.set_xlabel("\# Documents")
ax.set_ylabel("Relative throughput")
ax.set_xticks([ 500000, 1000000])
ax.set_xticklabels(["0.5M", "1M"])
ax.set_title("10\% U, 90\% S")

plt.legend()
plt.tight_layout()
plt.savefig("out/fig11a.png")
plt.show()
