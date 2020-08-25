import re
import custom_style
from custom_style import setup_columns,col,remove_chart_junk
import matplotlib.pyplot as plt
import sys
import numpy as np
from matplotlib.ticker import FuncFormatter
import math
from collections import defaultdict
from matplotlib.patches import Patch
from matplotlib.ticker import ScalarFormatter
import scipy.special

out_name = sys.argv[3] 
in_dory = sys.argv[1]
in_oram = sys.argv[2]

dory_color = custom_style.hash_colors[1]
oram_color =  custom_style.hash_colors[0]

plt.rcParams.update({'font.size':6})

dory = []
x = [2**j for j in range(10,21)]
with open(in_dory, 'r') as f:
    for i, line in enumerate(f):
        dory.append(float(line) / 1000.0)

oram = []
with open(in_oram, 'r') as f:
    for i, line in enumerate(f):
        oram.append(float(line) / 1000.0)
oram_x = [2**j for j in range(10, len(oram)+10)]

# N = 10000, n = 100
fig = plt.figure(figsize = (8,8))
ax = fig.add_subplot(111)
print(x)
#ax.loglog(x, semihonest, basex=2, basey=10, color=colors[0], linewidth=1, marker="o", linestyle="--", label="DORY (semihonest, 1 cluster)")
ax.loglog(x, dory, basex=2, basey=10, color=dory_color, linewidth=1, label=r"DORY")
ax.loglog(oram_x, oram, basex=2, basey=10, color=oram_color, linewidth=1, label="PathORAM baseline", linestyle="dashed")
#ax.stackplot(x, ySeconds[0], ySeconds[1], labels=labels, colors=colors)
#ax.stackplot(np.arange(10, 110, step=10), y[0], y[1], y[2], y[3], labels=labels, colors=colors)
ax.set_xlabel(r"\# Documents")
ax.set_ylabel("Update latency (s)")
#ax.set_yticks([100, 1000, 10000, 100000, 1000000])
#ax.set_yticklabels(["0.1", "1.0", "10.0", "100.0", "1000.0"])
ax.set_xticks([1024, 32768, 1048576])
ax.set_yticks([0.01, 1, 100, 10000])
#ax.set_xticks([1024, 32768, 1048576])
#ax.yaxis.set_major_formatter(ScalarFormatter())
#ax.set_xticks([0, 250, 500, 750, 1000])
#ax.set_xticklabels(["0", "250k", "500k", "750k", "1M"])
#ax.set_ylim([0,10])
#ax.set_xlim([10,110])
#ax.set_xticks(range(20,101,20))
#ax.set_yticks([0,2,4,6,8,10])
#handles, labels = ax.get_legend_handles_labels()
#handles.reverse()
#labels.reverse()
#ax.legend(handles, labels, bbox_to_anchor=(0., 1.02, 1., .102), loc='lower left', ncol=1, borderaxespad=0., fontsize=7,labelspacing=0)
#ax.legend(bbox_to_anchor=(-0.4,1.02, 1, 0.2), loc="lower left", labelspacing=0)
#ax.legend(bbox_to_anchor=(-0.3, 1.2, 0.9, .102), labelspacing=0)


#ax.spines['left'].set_position("zero")
#ax.spines['bottom'].set_position("zero")
remove_chart_junk(plt,ax,grid=True,below=True)

custom_style.save_fig(fig, out_name, [1.25, 1.2])
#custom_style.save_fig(fig, out_name, [1.5, 1.9])
#custom_style.save_fig(fig, out_name, [3.25, 1.8])
#plt.show()
