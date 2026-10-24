package main

// Stage 14: Probability, Statistics & Classical Machine Learning. The bridge
// between the maths of Stage 13 and the neural networks of Stage 15.

var dataGuides = map[int]StageGuide{
	14: {
		Overview: `Before neural networks, there is data. This stage teaches you to
describe data, to reason about uncertainty, and to build the classic
machine-learning models (regression, trees and clustering). Above all,
it teaches the most important skill in ML: evaluating honestly whether a
model actually works. Every lesson is measured: an A/A test that "finds"
an effect 37% of the time when you peek, a medical test that is wrong
five times in six, and a model that scores AUC 0.99 because the answer
leaked into its inputs.`,
		Outcomes: []string{
			"Summarise and visualise data without fooling yourself",
			"Tell real effects from luck with experiments and statistics",
			"Build, compare and honestly evaluate classic ML models",
			"Turn a test's accuracy and a base rate into the probability that matters, with Bayes' theorem",
			"Put an error bar on any statistic with the bootstrap",
			"Write regression, k-means and model metrics from scratch in C",
		},
		Glossary: []Term{
			{"Mean", "The average: add everything up and divide by how many there are."},
			{"Median", "The middle value once the data is sorted."},
			{"Standard deviation", "A measure of how spread out values are around the mean."},
			{"Correlation", "How strongly two quantities move together; it does not prove that one causes the other."},
			{"p-value", "How surprising your data would be if there were no real effect."},
			{"Feature", "An input variable a model uses, such as a house's size or age."},
			{"Overfitting", "When a model memorises its training data and fails on new data."},
			{"Precision", "Of the items a model flagged, the fraction that were truly positive."},
			{"Recall", "Of the truly positive items, the fraction the model caught."},
			{"Base rate", "How common something is before any test: the prevalence of a disease, the share of transactions that are fraud."},
			{"Prior and posterior", "What you believed before seeing the evidence, and what you should believe after it."},
			{"Statistical power", "The chance that an experiment detects an effect that is really there."},
			{"Bootstrap", "Estimating the uncertainty of a statistic by recomputing it on many resamples, drawn with replacement, of your own data."},
			{"Bias and variance", "Two ways a model errs: bias from being too simple to fit the pattern, variance from fitting the noise of one particular sample."},
			{"Regularisation", "A penalty on large weights that makes a model simpler, trading a little bias for much less variance."},
			{"AUC", "The area under the ROC curve: the probability that the model scores a random positive above a random negative."},
			{"Data leakage", "Information that would not exist at prediction time getting into a model's inputs, which makes its scores look better than they will be."},
		},
		Concepts: []Concept{
			{
				Name:    "Descriptive Statistics & Distributions",
				Summary: "Summarising data honestly, and seeing its shape.",
				Body: `The mean is the average. The median is the middle value, and it is
robust to outliers. The standard deviation measures spread. A
distribution describes how values are spread out. The normal (bell)
curve arises when many small effects add up (heights, measurement
errors). Skewed distributions such as incomes have long tails, where the
mean is pulled far above the median.

Always plot the data. Histograms and scatter plots reveal what summary
numbers hide: Anscombe's quartet is four datasets with the same means,
variances and correlation that look completely different when drawn.
Correlation measures how two variables move together, but correlation is
not causation.`,
				UnderTheHood: `Anscombe's quartet, recomputed with NumPy: four sets of 11 points.

  set   mean x  var x  mean y  var y   correlation  fitted line
   1     9.00   11.00   7.50   4.127      0.816     y = 3.00 + 0.500x
   2     9.00   11.00   7.50   4.128      0.816     y = 3.00 + 0.500x
   3     9.00   11.00   7.50   4.123      0.816     y = 3.00 + 0.500x
   4     9.00   11.00   7.50   4.123      0.817     y = 3.00 + 0.500x

Identical summaries; drawn, they are a cloud, a curve, a straight line
with one outlier, and a vertical stack with one far point.

Means and medians of incomes, simulated with a log-normal distribution
(the usual shape of incomes):

  49 people                    mean 42,507     median 37,766
  the same 49, plus one with   mean 20,041,657 median 38,334
  1,000,000,000
  1,000,000 simulated incomes  mean 48,139     median 35,003

One extra person moved the mean by a factor of 470 and the median by
1.5%. And in a skewed distribution most people are below average: 65.5%
of the million earned less than the mean.`,
				MentalModel: "Plot first, summarise second, and never trust a single number.",
				TryIt:       "Compute the mean and median of a list of salaries, then add one billionaire and compute them again.",
				Analogy: `When a billionaire walks into a café, the average wealth in the room
jumps to hundreds of millions, while the median customer is exactly as
well off as before. The median describes a typical person; the mean can
be hijacked by a single extreme value.`,
				Example: `Governments usually report median household income rather than the mean
for exactly this reason. Ice-cream sales and drownings are correlated,
because both rise in summer, but ice cream does not cause drowning.`,
				Exercises: trio(
					"For [2, 3, 3, 4, 100], compute the mean, the median and the mode. Which best describes a typical value?",
					"The mean is 22.4; the median and the mode are both 3.",
					"In C, read numbers from standard input and print their count, mean, median and standard deviation, using qsort for the median and Welford's running update for the variance. Test it on the list above and on a million values near 10⁹.",
					"Welford: for each x, n++, d = x − mean, mean += d / n, m2 += d · (x − mean); the variance is m2 / (n − 1).",
					"Track a personal metric for two weeks (sleep, steps or spending). Summarise it with suitable statistics and a chart, check whether it correlates with something else you tracked, and explain why the correlation might not be causal.",
					"A scatter plot plus numpy.corrcoef, or a spreadsheet's CORREL.",
				),
			},
			{
				Name:    "Bayes' Theorem & Base Rates",
				Summary: "How likely is it really, given the evidence? Start from how common it was.",
				Body: `Bayes' theorem turns "how likely is the evidence if the claim is true"
into "how likely is the claim, now that I have seen the evidence":

  P(ill | positive) = P(positive | ill) × P(ill) / P(positive)

The piece people forget is P(ill), the base rate. A test that is 99%
sensitive (it catches 99% of the ill) and 95% specific (it clears 95% of
the healthy) sounds decisive. But if only 1 person in 100 is ill, the 5%
false alarms among the 99 healthy outnumber the true positives, and most
positive results are wrong. Ignoring the base rate is called the base-rate
fallacy, and doctors, juries and engineers all fall for it.

The same arithmetic runs spam filters (naive Bayes multiplies the
evidence of each word), fraud alerts and security monitoring, where real
attacks are rare and false alarms bury the analysts.`,
				Diagram: `10,000 people, 1% ill, test 99% sensitive and 95% specific
   100 ill   ──▶  99 test positive  ─┐
 9,900 well  ──▶ 495 test positive  ─┴─▶ 594 positives, 99 of them ill
                                     P(ill | positive) = 99 / 594 ≈ 17%`,
				UnderTheHood: `The test above, 99% sensitive and 95% specific, at three base rates:

  prevalence   P(ill | positive test)
     20%            83.2%
      1%            16.7%
      0.1%           1.9%

The test is exactly as good in all three rows; only how common the
illness is changed. A simulation of 1,000,000 people at 1% prevalence
agreed: 59,695 tested positive, and 9,815 of them (16.4%) were really
ill, the rest false alarms from the much larger healthy group.

This is why screening programmes retest positives with a second,
different test: after one positive the probability is 16.7%, and that
becomes the new prior. A second independent positive raises it to
about 80%.`,
				MentalModel: "First ask how common it is; the evidence only adjusts that starting point.",
				TryIt:       "Redo the 10,000-people table with a test that is 90% sensitive and 99% specific, at 1% prevalence.",
				Analogy: `A smoke alarm in a house where fires are very rare: most of the times it
goes off, someone burnt the toast. The alarm is not broken; there is
simply far more toast than fire.`,
				Example: `In 2009 the US Preventive Services Task Force stopped recommending routine
mammograms for women in their forties partly for this reason: at low
prevalence most positive screens are false alarms, each one leading to
anxiety, biopsies and follow-up. Security teams face the same maths when
an intrusion detector with a 1% false-positive rate inspects a million
benign events a day.`,
				Exercises: trio(
					"A test is 98% sensitive and 97% specific, and 2% of the population has the condition. What is P(condition | positive)?",
					"Of 10,000: 200 ill, 196 positive; 9,800 well, 294 positive. 196 / 490 = 40%.",
					"In C, write a naive Bayes spam filter: count how often each word appears in a set of spam and non-spam messages, then classify new messages by adding log-probabilities, with add-one smoothing so an unseen word does not make a probability 0.",
					"Work in logarithms: multiplying hundreds of small probabilities underflows to 0, adding their logs does not.",
					"Find a real screening or alerting system (a medical test, your bank's fraud texts, a spam filter, a CI system's flaky-test alerts). Estimate its base rate and false-positive rate, and work out what fraction of its alerts are real.",
					"Ask how many alerts there are in a typical week and how many turned out to be real.",
				),
			},
			{
				Name:    "Statistical Inference & Experiments",
				Summary: "Telling real effects from luck.",
				Body: `A sample is a subset of a population, so estimates from samples are
uncertain. The central limit theorem says that averages of samples are
approximately normally distributed, which lets us compute confidence
intervals: ranges likely to contain the true value.

Hypothesis testing asks: if there were no real effect, how surprising is
what we observed? The p-value is that probability. A small p-value
suggests a real effect, but it does not measure how big or important the
effect is. Randomised controlled experiments (A/B tests) are the gold
standard for showing cause and effect. Common pitfalls are tiny samples,
stopping as soon as the result looks good, and testing many things but
reporting only the winner (p-hacking).`,
				UnderTheHood: `Luck, measured. With a fair coin:

  7 or more heads in 10 flips      17.2%
  60 or more in 100                2.8%
  600 or more in 1,000             0.000000014%

The same 60% of heads goes from unremarkable to impossible as the
sample grows. Then 4,000 simulated A/A tests (both groups convert at
exactly 10%, so any "effect" is luck), with 10,000 users per group:

  test once, at the end                  4.7% false positives
  check after every 100 users, stop      37.5% false positives
  as soon as p < 0.05

Peeking turned the promised 5% error rate into 37.5%. With 20 metrics
checked independently, the chance that at least one shows p < 0.05 by
luck is 64%. And to detect a real lift from 10% to 11% conversion:

  users per group    1,000   3,000   8,000   15,000   20,000
  detected            11%     25%     54%      82%      91%

Most small A/B tests are too small to see the effects they are looking
for: decide the sample size before you start.`,
				MentalModel: "Before believing a difference, ask how often luck alone would produce it.",
				TryIt:       "Flip a simulated fair coin 100 times, then 1,000 times, repeatedly. How often do you see 60% or more heads?",
				Analogy: `A fertiliser trial: plant two randomly chosen groups of seeds, one with
fertiliser and one without. If the treated plants grow taller,
randomisation means the only systematic difference was the fertiliser.
But with only a few plants, luck could easily explain the gap.`,
				Example: `Technology companies run thousands of A/B tests a year. At Microsoft's
Bing, a small change to how ad headlines were displayed raised revenue by
about 12%, after the idea had sat in the backlog for months. Clinical
trials for vaccines and drugs use the same randomised design.`,
				Exercises: trio(
					"A friend flips a coin 10 times and gets 7 heads. Is the coin biased? Estimate how likely 7 or more heads is with a fair coin.",
					"About 17%, so not surprising at all.",
					"Simulate an A/B test in which version B truly converts at 11% and version A at 10%. How many users per group do you need before the test detects the difference most of the time?",
					"Repeat the simulated experiment many times at each sample size and count how often p < 0.05.",
					"Run a real mini-experiment (two message subject lines, or two study techniques on alternating days). Randomise, collect data for two weeks, analyse it with a confidence interval, and state its limitations honestly.",
					"Decide the metric and the sample size before you start.",
				),
			},
			{
				Name:    "The Bootstrap & Resampling",
				Summary: "An error bar for any statistic, computed from your own data.",
				Body: `Formulas for confidence intervals exist for the mean and a few other
statistics, but not for the median, a ratio, a model's AUC or the
difference between two models. The bootstrap (Efron, 1979) needs no
formula: treat your sample as the population, draw a new sample of the
same size from it with replacement, compute the statistic, and repeat
thousands of times. The spread of those values is the uncertainty.

The 2.5th and 97.5th percentiles of the resampled statistics give a 95%
interval (the percentile bootstrap). Permutation tests use the same idea
for hypothesis tests: shuffle the group labels many times to see how big
a difference luck produces, and compare with the real one.

The bootstrap cannot invent information. It assumes the sample is
representative and the observations independent; with 10 data points or
with time series it can be badly wrong.`,
				Diagram: `your sample (n values) ──draw n with replacement──▶ resample 1 ─▶ median₁
                       ──draw n with replacement──▶ resample 2 ─▶ median₂
                                    …  10,000 times  …
            the 2.5% and 97.5% points of all the medians = a 95% interval`,
				UnderTheHood: `200 simulated incomes (log-normal, true median 35,000), 10,000 bootstrap
resamples:

  sample median                 35,552
  bootstrap 95% interval        30,094 to 41,079

Is the interval honest? Repeating the whole experiment on 500 fresh
samples, each with its own bootstrap interval, the interval contained
the true median 94.6% of the time: close to the 95% it claims.

A bootstrap resample of 10,000 values contained 63.2% of the original
values at least once (the rest were repeats), exactly the 1 − 1/e that
theory predicts. Random forests use this: each tree is trained on a
bootstrap sample, and the 36.8% of rows a tree never saw become a free
test set for it (the "out-of-bag" estimate).`,
				MentalModel: "Resample your data to see how much your answer would wobble.",
				TryIt:       "Bootstrap the median of 20 numbers 1,000 times with numpy.random.choice, and plot a histogram of the medians.",
				Analogy: `Tasting a pot of soup with many small spoonfuls, putting each back: the
variation between spoonfuls tells you how much one spoonful can be
trusted, without cooking a second pot.`,
				Example: `Machine-learning papers report bootstrap intervals on benchmark scores to
show whether one model really beats another, and economists and
biostatisticians use it wherever no formula exists. Random forests
(Breiman, 2001) are built on bagging, "bootstrap aggregating": every
tree is trained on its own bootstrap sample of the data.`,
				Exercises: trio(
					"Why must a bootstrap sample be drawn with replacement, and be the same size as the original?",
					"Without replacement, every resample is the original in a new order, so the statistic never changes.",
					"In C, compute a bootstrap 95% interval for the median of numbers read from standard input, using a small random-number generator of your own (xorshift64 is ten lines) and qsort, with 10,000 resamples.",
					"xorshift64: x ^= x << 13; x ^= x >> 7; x ^= x << 17; then x % n picks an index.",
					"Take two models' predictions on the same test set (or two versions of your own program's timings) and decide with a bootstrap of the difference whether one is really better, or the gap is luck.",
					"Resample the test items, not the two scores: the same items for both models in each resample.",
				),
			},
			{
				Name:    "Linear & Logistic Regression",
				Summary: "The simplest models that learn from data, and still widely used.",
				Body: `Linear regression fits a straight line (or plane), y = w·x + b, to
predict a number, choosing the weights that minimise the squared error.
It can be solved exactly, or by gradient descent (Stage 13). Logistic
regression predicts the probability of a yes/no outcome by passing
w·x + b through the sigmoid function, and it is trained with
cross-entropy.

Both are fast and interpretable (each weight shows a feature's
influence), and both make strong baselines. Feature engineering
(choosing and transforming the inputs) often matters more than the
choice of model. Regularisation (L1, L2) penalises large weights to
reduce overfitting.`,
				Diagram: `price │                     •   ╱
      │                 •   ╱  •
      │            •   ╱ •          fitted line: price = w·size + b
      │        •  ╱  •
      │     • ╱
      └──────────────────────────── size`,
				UnderTheHood: `500 simulated houses, price = 1,500·size − 800·age + 50,000 + noise, size
in m² (30 to 250) and age in years (0 to 80). The exact least-squares
solution gives size 1,507.1, age −744.8, intercept 46,215. Gradient
descent, stopping when the loss improves by less than one part in a
million:

  raw features, step 10⁻⁴      diverged after 863 steps
  raw features, step 2·10⁻⁵    stopped after 190,781 steps, intercept 15,956
  raw features, step 10⁻⁵      stopped after 743 steps, intercept 42
  scaled features, step 0.1    85 steps, size 1,506.9, age −744.7,
                                intercept 46,210

Unscaled, the size feature (in the hundreds) dominates the gradient, so
any step small enough not to blow up crawls along the others; the 743-step
run "converged" with an intercept of 42 instead of 46,215. Scaling each
feature to mean 0 and standard deviation 1 fixed it in 85 steps.

Logistic regression on perfectly separable data has no best weight:
without regularisation the weight grew from 3.6 after 100 steps to 12.8
after 100,000, and kept growing. With an L2 penalty of 0.01 it stopped at
2.5 from the first hundred steps on.`,
				MentalModel: "Start with the simplest model, and beat it before reaching for a complex one.",
				TryIt:       "Fit a line to 10 (x, y) points with hand-written gradient descent and compare the result with numpy.polyfit.",
				Analogy: `Guessing a house's price from its size: you mentally draw a trend line
through past sales. Logistic regression is like a doctor's rule of thumb
that turns several risk factors into a single "chance of illness".`,
				Example: `Credit scoring has used logistic regression for decades, partly because
regulators require decisions that can be explained. Retailers forecast
demand with regression models, and many hospital early-warning scores are
built on logistic models.`,
				Exercises: trio(
					"The line y = 2x + 1 predicts y for x = 3, but the real observation is 8. What are the error and the squared error?",
					"The prediction is 7, so the error is 1 and the squared error is 1.",
					"In C, fit price = w·size + b by gradient descent on 200 (size, price) pairs, first with raw sizes and then with standardised ones, and print the number of steps and the final w and b for each. Compare with the exact formula w = Sxy / Sxx.",
					"Standardise with z = (x − mean) / sd, then convert the weight back: w_raw = w_z / sd.",
					"Predict used-car prices from a public dataset (age, mileage, brand). Report the average error in currency and explain which features matter most. Then use logistic regression to predict whether a car sells above the median price.",
					"One-hot encode categorical features such as brand.",
				),
			},
			{
				Name:    "Bias, Variance & Regularisation",
				Summary: "Too simple misses the pattern; too flexible memorises the noise.",
				Body: `A model's error on new data has two sources besides plain noise. Bias:
the model is too simple to represent the pattern, so it is wrong in the
same way on every training set (a straight line through a curve).
Variance: the model is so flexible that it fits the noise of the sample
it was given, so it changes wildly from one training set to the next.

Making a model more flexible lowers bias and raises variance; the test
error is lowest somewhere in between. Training error alone always
improves with flexibility and cannot find that point, which is why you
choose a model's complexity with a validation set or cross-validation.

Regularisation keeps a flexible model but penalises large weights: ridge
(L2) adds λ·Σw² to the loss and shrinks every weight, lasso (L1) adds
λ·Σ|w| and drives some to exactly zero. More data lowers variance too:
the best cure for overfitting is often simply more examples.`,
				Diagram: `error │╲                         ╱  test error
      │ ╲                     ╱
      │  ╲___            ____╱
      │      ╲__________╱        ◀ the sweet spot
      │  ╲___
      │      ╲________________   training error
      └──────────────────────── model flexibility
        high bias          high variance`,
				UnderTheHood: `Polynomials fitted by least squares to 20 noisy points from a sine curve
(noise of standard deviation 0.3, so the best possible error is 0.09),
median of 200 repetitions, test error on 500 fresh points:

  degree    training error   test error
     1          0.270          0.312     too simple: high bias
     3          0.072          0.112     about right
     5          0.058          0.130
     9          0.040          1.131
    15          0.036         39.157     memorised the noise

The training error fell at every step; the test error was best at
degree 3 and exploded by degree 15. Ridge regularisation rescued the
degree-15 model:

  ridge λ    10⁻⁶     10⁻³     10⁻¹
  test       0.157    0.137    0.210

A tiny penalty took the test error from 39 to 0.14; too large a penalty
made the model too stiff again (0.210).`,
				MentalModel: "Choose complexity on data the model never trained on; training error always says more is better.",
				TryIt:       "Fit polynomials of degree 1, 3 and 15 to 20 noisy sine points with numpy.polyfit and plot them over the true curve.",
				Analogy: `A student revising for an exam. One learns a single rule of thumb and
gets every subtle question wrong (bias). Another memorises last year's
paper word for word and is lost when the questions change (variance).
The one who understands the material does well on a paper they have
never seen.`,
				Example: `Every serious ML pipeline tunes regularisation strength (weight decay in
neural networks, tree depth in gradient boosting, λ in ridge regression)
by cross-validation. In 2007–2009 the Netflix Prize teams found that
heavily regularised models beat larger unregularised ones on the hidden
test set.`,
				Exercises: trio(
					"A model scores 99% on its training data and 70% on new data. Is the problem bias or variance, and name two things that would help.",
					"Variance (overfitting): more data, a simpler model, or regularisation.",
					"In C, fit polynomials of degree 1 to 12 to 20 noisy points by solving the normal equations with Gaussian elimination (Stage 13), and print the training and test error for each degree.",
					"Scale x to [−1, 1] first: the powers of a large x make the system terribly conditioned.",
					"Take a real dataset, split off a test set, and choose a model's complexity (tree depth, polynomial degree or λ) with 5-fold cross-validation on the rest. Report the chosen value and the final test score, touched only once.",
					"scikit-learn's GridSearchCV does the folds; keep the test set out of it.",
				),
			},
			{
				Name:    "Trees, Ensembles & Clustering",
				Summary: "Models that ask questions, vote together, or find groups on their own.",
				Body: `A decision tree predicts by asking a series of yes/no questions about the
features ("income > 40k?"), learned so that each split separates the
data as cleanly as possible. Trees are easy to explain but overfit
easily.

Ensembles combine many trees. Random forests average trees trained on
random subsets of the data. Gradient-boosted trees (XGBoost, LightGBM)
add trees one at a time, each correcting the previous ones' errors, and
remain a top choice for tabular data. Clustering is unsupervised: k-means
groups points around k centres without any labels, and
k-nearest-neighbours classifies a point by looking at its closest
examples.`,
				Diagram: `             [ income > 40k ? ]
              yes ╱        ╲ no
      [ has debt ? ]        deny
       yes ╱    ╲ no
       review   approve`,
				UnderTheHood: `scikit-learn on a synthetic task (20 features, 6 informative, 10% of
labels flipped at random), 2,000 rows to train and 2,000 to test:

  model                       train    test
  logistic regression         0.643    0.619
  decision tree, unlimited    1.000    0.767   (225 leaves)
  decision tree, depth 5      0.740    0.704
  random forest, 300 trees    1.000    0.870
  gradient boosting, 300      0.976    0.849

The unlimited tree memorised every training row, noise included. A
forest of 300 such trees, each grown on a bootstrap sample, scored 0.870
on new data: averaging cancels much of each tree's variance.

k-means on 3,000 points in 8 clear blobs, 50 runs from different random
starts:

  random starting centres   best inertia 2,200, worst 16,835;
                            8 of 50 runs within 1% of the best
  k-means++ starts          best 2,200, worst 2,929; 35 of 50

k-means only finds a local optimum, so libraries run it several times
from k-means++ starts and keep the best. And it measures plain distance:
clustering people by (age, income) found the two real age groups 53% of
the time (chance) with raw features, where income's spread swamped age,
and 100% after scaling both.`,
				MentalModel: "One tree explains; a forest predicts.",
				TryIt:       "Train a decision tree on the Iris dataset with scikit-learn and print its rules.",
				Analogy: `The game "20 Questions" is a decision tree. A random forest is asking 100
friends to play and taking the majority answer. Clustering is sorting a
pile of mixed socks into groups by colour without being told what the
colours are.`,
				Example: `Banks detect fraud, and insurers price policies, with gradient-boosted
trees, which have also won a large share of Kaggle competitions on
tabular data. Marketing teams use clustering to find customer segments.`,
				Exercises: trio(
					"Draw a decision tree, at most 3 questions deep, that decides whether to take an umbrella. Then describe a case where your tree gives the wrong answer.",
					"Consider the forecast, the sky and the season.",
					"In C, implement k-means for 2D points read from standard input, starting from the first k points, and print each centre with its number of points. Then run it on three blobs for k = 2, 3 and 5.",
					"Repeat: assign each point to its nearest centre, then move each centre to the mean of its points, until no assignment changes.",
					"On a public customer-churn or loan-default dataset, train a decision tree, a random forest and a gradient-boosted model, compare them fairly on held-out data, and explain the top 3 features behind the best model's predictions.",
					"Use scikit-learn's feature_importances_ or permutation importance.",
				),
			},
			{
				Name:    "Evaluating Models Honestly",
				Summary: "Accuracy can lie; measure what matters.",
				Body: `Always evaluate on data the model has never seen. Split the data into
training, validation (for tuning) and test sets, or use k-fold
cross-validation. Accuracy misleads on imbalanced data: if 99% of
transactions are legitimate, a model that always says "legitimate" is
99% accurate and completely useless.

Use a confusion matrix, precision (of the items flagged, how many are
right), recall (of the real positives, how many were caught) and the F1
score, and choose the trade-off your application needs. Watch for data
leakage, where information from the future or from the answer itself
sneaks into the features. Leakage is the most common reason a model
looks great in development and then fails in production.`,
				Diagram: `              predicted +       predicted −
actually + │ true pos (TP)  │ false neg (FN) │ recall    = TP/(TP+FN)
actually − │ false pos (FP) │ true neg (TN)  │ precision = TP/(TP+FP)`,
				UnderTheHood: `A synthetic fraud-like task: 1% positives, 100 of them in a test set of
10,000.

  model                        accuracy  precision  recall  flagged
  always "no"                    0.990       –       0.00       0
  logistic, threshold 0.50       0.990     0.40      0.08      20
  logistic, threshold 0.20       0.988     0.36      0.30      83
  logistic, threshold 0.05       0.965     0.17      0.65     378

Accuracy barely moved while recall went from 0 to 0.65: the threshold,
not the model, decided how many frauds were caught and how many alerts
the analysts must read.

Leakage, on purpose: the same random forest with honest features scored
AUC 0.894; adding one feature that secretly encoded the answer (plus
noise) gave 0.992, a score that collapses in production, where that
feature does not exist yet.

Reusing a test set: on purely random labels, the best of 1,000 random
models chosen by their test score reached 0.570 accuracy on that test
set, and 0.516 (chance) on fresh data. Every time you look at the test
set to choose, it becomes a little more training data.`,
				MentalModel: "A model is only as good as the honesty of its evaluation.",
				TryIt:       "Build a 'model' that always predicts the majority class, and compute its accuracy, precision and recall.",
				Analogy: `A smoke alarm. High recall means it never misses a fire; high precision
means it rarely goes off because of burnt toast. You would rather put up
with some toast alarms than miss a fire. The right balance depends on
what each kind of mistake costs.`,
				Example: `Cancer screening favours high recall (do not miss cases) and follows up
with more precise tests, while spam filters favour precision (do not
lose real email). In 2021 Zillow shut down its home-buying business after
its pricing models overestimated home values and it lost hundreds of
millions of dollars: good offline scores are not the same as real-world
performance.`,
				Exercises: trio(
					"A fraud model flags 100 transactions, of which 20 are real fraud. There were 50 frauds in total. Compute the precision and the recall.",
					"Precision = 20/100 = 0.2; recall = 20/50 = 0.4.",
					"On an imbalanced dataset (such as credit-card fraud from Kaggle), train a classifier, plot precision and recall as you vary the decision threshold, and choose a threshold for a bank that can review 100 alerts a day.",
					"predict_proba gives scores; sweep the threshold from 0 to 1.",
					`Create data leakage on purpose: predict whether customers will cancel using a feature such as "cancellation_date is not empty", observe the suspiciously perfect score, then remove it and report the honest score. Write a checklist for catching leakage in future projects.`,
					"Ask of every feature: would I know this at the moment of prediction?",
				),
			},
		},
		Resources: []Resource{
			{"Book", "An Introduction to Statistical Learning (free)", "https://www.statlearning.com/", "The friendliest serious ML textbook, with Python and R labs."},
			{"Video", "StatQuest with Josh Starmer", "https://statquest.org/", "Clear, cheerful explanations of statistics and ML."},
			{"Course", "Machine Learning Specialization (Andrew Ng)", "https://www.coursera.org/specializations/machine-learning-introduction", "The classic beginner ML course, updated."},
			{"Site", "scikit-learn user guide", "https://scikit-learn.org/stable/user_guide.html", "Practical documentation for every classic model."},
			{"Course", "Kaggle Learn", "https://www.kaggle.com/learn", "Short, free, hands-on courses with real datasets."},
			{"Tool", "Seeing Theory", "https://seeing-theory.brown.edu/", "Beautiful interactive visualisations of probability and statistics."},
		},
		Blueprints: []Blueprint{
			{"ML From Scratch", "Implement linear regression, logistic regression, a decision tree and k-means using only NumPy, and compare them with scikit-learn.",
				[]string{"Linear regression", "Logistic regression", "Decision tree", "k-means", "Compare with scikit-learn"}},
			{"End-to-End Prediction Project", "Pick a real public dataset and question, then clean, explore, model, evaluate honestly and present your findings.",
				[]string{"Question and dataset", "Cleaning and exploration", "Baseline model", "Better model", "Honest evaluation and write-up"}},
			{"A/B Testing Toolkit", "A small library and command-line tool for planning sample sizes and analysing A/B tests, with confidence intervals.",
				[]string{"Sample-size calculator", "Two-proportion test", "Confidence intervals", "Simulation-based checks"}},
		},
		Quiz: []Question{
			{"When is the median better than the mean?", "When the data has outliers or is skewed, such as incomes or house prices."},
			{"What does a p-value of 0.03 mean?", "If there were no real effect, data at least this extreme would appear about 3% of the time. It is not the probability that the effect is real."},
			{"Why is accuracy misleading on imbalanced data?", "Always predicting the majority class scores highly while catching none of the rare cases."},
			{"What is data leakage?", "Information that would not be available at prediction time sneaking into the training features, which inflates the scores."},
			{"A test is 99% sensitive and 95% specific, and 1% of people are ill. Roughly what fraction of positive results are real?", "About 17%: the 5% false alarms among the 99% who are well outnumber the true positives."},
			{"Why does checking an A/B test every day and stopping at the first p < 0.05 give wrong answers?", "Every look is another chance for luck to cross the line: peeking every 100 users raised a 5% false-positive rate to about 37%."},
			{"How does the bootstrap estimate the uncertainty of a median?", "Resample the data with replacement many times, compute the median of each resample, and read the spread (for example the 2.5th and 97.5th percentiles)."},
			{"A model's training error keeps falling as you make it more flexible. Why can you not use it to choose the flexibility?", "Training error always improves with flexibility, even while the model memorises noise; only error on held-out data shows where test performance peaks."},
			{"Why should features be scaled before gradient descent or k-means?", "A feature with a much larger range dominates the gradient or the distance, so descent crawls or diverges and k-means ignores the other features."},
			{"Why does a random forest generalise better than a single unlimited decision tree?", "Each tree overfits differently on its own bootstrap sample; averaging many cancels much of that variance."},
			{"What goes wrong if you pick the best of many models by their score on the test set?", "The test set becomes part of the selection, so the score is optimistic: on random labels the best of 1,000 random models scored 57% there and 52% on fresh data."},
		},
	},
}
