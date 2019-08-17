
# Analysing data for QSM

## Finding which growth context has best spherical approximation

Each growth context creates new set of open nodes on each round.
We need measurements that shows the closeness of this set to a sphere, what is the radius of the sphere, how distributed (far, amount) from the sphere the cloud of points are.
It'll be great to go all the way to 128 steps per contexts. From initial measurements 128 steps is taking way too much time.
So, optimization is needed before.

In the mean time we can run theses measurements all the way to 8*6 = 48 steps.
For each growth context and each steps we need a new line of data with:
   - Growth Context ID
   - Step number: n
   - Number of open nodes
   - Avg, mean distance from the center
   - Histogram using step of 1 integer sequencing around the average

Then data analysis should provide some following numbers for each growth context + step number:
- ratio from number of nodes and n^2,
- ratio between avg distance and n,
- number quality of the histogram (standard deviation) 

Then for each growth context judge the standard deviation from the 2 ratios (removing the first 6 steps) since a stable ratios over steps is important.
Then identified patterns between growth context as many will behave in a very similar way.

