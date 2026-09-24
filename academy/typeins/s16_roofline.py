# ROOFLINE -- how fast can a language model possibly talk?
# At batch size 1, every generated token reads every weight once,
# so tokens per second <= memory bandwidth / model size.

models = [("7B, 16-bit", 7e9, 2.0), ("7B, 4-bit", 7e9, 0.5), ("70B, 4-bit", 70e9, 0.5)]
devices = [("laptop, 100 GB/s", 100e9), ("GPU, 1000 GB/s", 1000e9)]

for name, params, bytes_per_param in models:
    size_gb = params * bytes_per_param / 1e9
    ceilings = ", ".join(f"{dev}: {bw / (params * bytes_per_param):5.1f} tok/s"
                         for dev, bw in devices)
    print(f"{name:11} ({size_gb:4.1f} GB)  {ceilings}")
