## Accelerated PyTorch training on Mac

## Metal acceleration

PyTorch uses the Metal Performance Shaders (MPS) backend for GPU acceleration. This MPS backend 
extends the PyTorch framework, providing scripts and capabilities to set up and run operations on 
Mac. The MPS framework optimizes compute performance with kernels that are fine-tuned for the 
unique characteristics of each Metal GPU family. The `mps` device maps machine learning 
computational graphs and primitives on the MPS Graph framework and tuned kernels provided by MPS 
— delivering optimal performance out of the box. Custom Metal kernels are also supported for 
developers wanting additional flexibility.

## Requirements

The following requirements apply to the latest stable release of PyTorch (2.11.0):

- Mac computer with Apple silicon
- macOS 14.0 or later
- Python 3.10 or later
- Xcode command-line tools: `xcode-select --install`

## Get started

#### 1\. Installation

Stable (PyTorch 2.11.0):

```
pip3 install torch torchvision torchaudio
```

Nightly (latest MPS support):

```
pip3 install --pre torch torchvision torchaudio --extra-index-url 
https://download.pytorch.org/whl/nightly/cpu
```

#### 2\. Building from source

For instructions on building PyTorch with MPS support from source, refer to the [official PyTorch 
GitHub repository](https://github.com/pytorch/pytorch).

#### 3\. Verify

You can verify `mps` support using a simple Python script:

```
import torch
if torch.backends.mps.is_available():
    mps_device = torch.device("mps")
    x = torch.ones(1, device=mps_device)
    print (x)
else:
    print ("MPS device not found.")
```

The output should show:

```
tensor([1.], device='mps:0')
```

## Feedback

The MPS backend is in the beta phase, and we’re actively addressing issues and fixing bugs. To 
report an issue, use the [GitHub issue 
tracker](https://github.com/pytorch/pytorch/issues?q=is%3Aissue+is%3Aopen+module%3A+mps) with the 
label “module: mps”.

## Resources

[PyTorch installation page](https://pytorch.org/get-started/locally/)  
[PyTorch documentation on MPS backend](https://pytorch.org/docs/master/notes/mps.html)  
[Add a new PyTorch operation to MPS 
backend](https://github.com/pytorch/pytorch/wiki/MPS-Backend#adding-op-for-mps-backend)  
[PyTorch performance profiling using MPS 
profiler](https://github.com/pytorch/pytorch/wiki/MPS-Backend#pytorch-performance-profiling-using-mp
s-profiler)
