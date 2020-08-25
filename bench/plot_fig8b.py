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

oram = []
with open("ref/latency_search_oram.dat", 'r') as f:
    for i, line in enumerate(f):
        oram.append(float(line))
oram_x = [2**j for j in range(10, len(oram)+10)]

malicious_1 = [x / 1000.0 for x in malicious_1]
malicious_2 = [x / 1000.0 for x in malicious_2]
malicious_4 = [x / 1000.0 for x in malicious_4]
oram = [x / 1000.0 for x in oram]

# N = 10000, n = 100
fig = plt.figure(figsize = (8,8))
ax = fig.add_subplot(111)
ax.loglog(malicious_1_x, malicious_1, basex=2, basey=10, color=dory_1_color, linewidth=1, label=r"DORY ($p=1$)")
ax.loglog(malicious_2_x, malicious_2, basex=2, basey=10, color=dory_2_color, linewidth=1, label=r"DORY ($p=2$)")
ax.loglog(malicious_4_x, malicious_4, basex=2, basey=10, color=dory_4_color, linewidth=1, label=r"DORY ($p=4$)")
ax.loglog(oram_x, oram, basex=2, basey=10, color=oram_color, linewidth=1, label="PathORAM baseline", linestyle="dashed")
ax.set_xlabel(r"# Documents")
ax.set_ylabel("Search latency (s)")
ax.set_yticks([0.1, 10.0, 1000.0])
ax.set_xticks([1024, 32768, 1048576])

plt.legend()
plt.tight_layout()
plt.savefig("out/fig8b.png")
plt.show()
