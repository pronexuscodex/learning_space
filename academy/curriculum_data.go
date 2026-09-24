package main

// Stage 14: Probability, Statistics & Classical Machine Learning. The bridge
// between the maths of Stage 13 and the neural networks of Stage 15.

var dataGuides = map[int]StageGuide{
	14: {
		Overview: `Before neural networks, there is data. This stage teaches you to
describe data, to reason about uncertainty, and to build the classic
machine-learning models (regression, trees and clustering). Above all,
it teaches the most important skill in ML: evaluating honestly whether a
model actually works.`,
		Outcomes: []string{
			"Summarise and visualise data without fooling yourself",
			"Tell real effects from luck with experiments and statistics",
			"Build, compare and honestly evaluate classic ML models",
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
					"Load a public dataset (from Kaggle or your national statistics office), compute the mean, median and standard deviation of three columns, and plot histograms. Describe each distribution's shape in one sentence.",
					"pandas describe() and matplotlib hist().",
					"Track a personal metric for two weeks (sleep, steps or spending). Summarise it with suitable statistics and a chart, check whether it correlates with something else you tracked, and explain why the correlation might not be causal.",
					"A scatter plot plus numpy.corrcoef.",
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
					"Implement linear regression with gradient descent from scratch (no ML libraries), fit it to house sizes and prices, and plot both the fitted line and the loss curve.",
					"Scale the features first, or use a very small learning rate.",
					"Predict used-car prices from a public dataset (age, mileage, brand). Report the average error in currency and explain which features matter most. Then use logistic regression to predict whether a car sells above the median price.",
					"One-hot encode categorical features such as brand.",
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
					"Implement k-means from scratch, run it on 2D points generated from three blobs, and plot the result for k = 2, 3 and 5.",
					"Repeat: assign each point to its nearest centre, then move each centre to the mean of its points.",
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
				Diagram: `                predicted +      predicted −
actually +  │  true pos (TP)  │ false neg (FN) │   recall    = TP / (TP + FN)
actually −  │  false pos (FP) │ true neg (TN)  │   precision = TP / (TP + FP)`,
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
		},
	},
}
