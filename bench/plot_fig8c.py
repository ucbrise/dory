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

plt.rcParams.update({'font.size':22})

malicious_1 = []
with open("out/latency_search_dory_1.dat", 'r') as f:
    for i, line in enumerate(f):
        malicious_1.append(float(line))
malicious_1_x = [2**j for j in range(10, len(malicious_1)+10)]

malicious_2 = []
with open("out/latency_search_dory_2.dat", 'r') as f:
    for i, line in enumerate(f):
        malicious_2.append(float(line))
malicious_2_x = [2**j for j in range(10, len(malicious_2)+10)]

malicious_4 = []
with open("out/latency_search_dory_4.dat", 'r') as f:
    for i, line in enumerate(f):
        malicious_4.append(float(line))
malicious_4_x = [2**j for j in range(10, len(malicious_4)+10)]

malicious_1 = malicious_1[5:]
malicious_2 = malicious_2[5:]
malicious_4 = malicious_4[5:]
malicious_1 = [j / 1000.0 for j in malicious_1]
malicious_2 = [j / 1000.0 for j in malicious_2]
malicious_4 = [j / 1000.0 for j in malicious_4]
malicious_1_x = malicious_1_x[5:]
malicious_2_x = malicious_2_x[5:]
malicious_4_x = malicious_4_x[5:]

# N = 10000, n = 100
fig = plt.figure(figsize = (8,8))
ax = fig.add_subplot(111)
ax.plot(malicious_1_x, malicious_1, color=dory_1_color, linewidth=1, label=r"DORY ($p=1$)")
ax.plot(malicious_2_x, malicious_2, color=dory_2_color, linewidth=1, label=r"DORY ($p=2$)")
ax.plot(malicious_4_x, malicious_4, color=dory_4_color, linewidth=1, label=r"DORY ($p=4$)")
ax.set_xlabel(r"# Documents")
ax.set_ylabel("Search latency (s)")
ax.set_yticks([0,1,2,3])
ax.set_xticks([0, 500000, 1000000])
ax.set_xticklabels(["0", "0.5M", "1M"])

ax.spines['left'].set_position("zero")
ax.spines['bottom'].set_position("zero")

plt.legend()
plt.tight_layout()
plt.savefig("out/fig8c.png")
plt.show()
