# ConsAlign

[![Build Status](https://travis-ci.org/kentwait/consalign.svg?branch=master)](https://travis-ci.org/kentwait/consalign)
[![Coverage Status](https://coveralls.io/repos/github/kentwait/consalign/badge.svg?branch=master)](https://coveralls.io/github/kentwait/consalign?branch=master)

Computes the consensus alignment by comparing alignment patterns
of global, local and sub-global multiple sequence alignments generated 
by the alignment program [MAFFT][1].

## Quickstart

### Align a set of nucleotide or protein sequences in a single FASTA file

    consalign input.fa > output.aln

### Align multiple FASTA files located in a single folder

    consalign -batch path/to/folder -outdir path/to/save/alignments

## Background

ConsAlign uses the multiple alignment program [MAFFT][1] to create three
multiple sequence alignments from one set of unaligned sequences using
three different alignment strategies - global alignment based on the
Needleman–Wunsch algorithm, local alignment based on the
Smith–Waterman algorithm, and local with affine-gap penalty scoring.

Even if these alignments were based on the same set of sequences, the
result could be vastly different. Thus, it is important to identify sites
that align robustly regardless of the alignment model used. We call
robustly aligning sites "consistent sites" and sites which exhibhttps://github.com/kentwait/consalign/releases/download/v1.0.1/consalign_linux_amd64it mixed
patterns "inconsistent sites".

## Method

ConsAlign identifies consistent sites by comparing the all alignment patterns
generated by the three strategies. It marks sites as "consistent" if the
particular alignment is rendered by all the given strategies. If the
particular alignment is not obeserved across all the alignment strategies
used, that site is marked "inconsistent". ConsAlign creates a marker sequence
at the head of the generated FASTA-formatted alignment to mark the status of
each site.

## Example - consistent sites

Here is an unaligned set of 5 sequences. Our aim is to find out if different
alignment strategies will affect the resulting alignment.

#### Unaligned

    >mel01
    gtaagtgtacacattatttccgatgtgggccttttgacgacaaaagaaatttatag
    >mel02
    gtaagtgtacacattatttccgatgtgggccttttgacgacaaaagaaatttatag
    >sim
    gtaagtgtacacattatttcggatgtgggtcttttgacgacaaagacatttatag
    >yak
    gtatgtgtacacgttatttctaatgtgaaacttttaacgacgaagacatttctag
    >ere
    gtaagtgtacacgttatttctaatgtgaaacttttgacgacaaaagacatttatag

The following are results of iterative global (G-INSI), local (L-INSI), and
affine-gap penalty (E-INSI) alignments.

#### Global alignment

    >mel01
    gtaagtgtacacattatttccgatgtgggccttttgacgacaaaagaaatttatag
    >mel02
    gtaagtgtacacattatttccgatgtgggccttttgacgacaaaagaaatttatag
    >sim
    gtaagtgtacacattatttcggatgtgggtcttttgacgac-aaagacatttatag
    >yak
    gtatgtgtacacgttatttctaatgtgaaacttttaacgac-gaagacatttctag
    >ere
    gtaagtgtacacgttatttctaatgtgaaacttttgacgacaaaagacatttatag

#### Local alignment

    >mel01
    gtaagtgtacacattatttccgatgtgggccttttgacgacaaaagaaatttatag
    >mel02
    gtaagtgtacacattatttccgatgtgggccttttgacgacaaaagaaatttatag
    >sim
    gtaagtgtacacattatttcggatgtgggtcttttgacgac-aaagacatttatag
    >yak
    gtatgtgtacacgttatttctaatgtgaaacttttaacgac-gaagacatttctag
    >ere
    gtaagtgtacacgttatttctaatgtgaaacttttgacgacaaaagacatttatag

#### Affine-gap alignment

    >mel01
    gtaagtgtacacattatttccgatgtgggccttttgacgacaaaagaaatttatag
    >mel02
    gtaagtgtacacattatttccgatgtgggccttttgacgacaaaagaaatttatag
    >sim
    gtaagtgtacacattatttcggatgtgggtcttttgacgac-aaagacatttatag
    >yak
    gtatgtgtacacgttatttctaatgtgaaacttttaacgac-gaagacatttctag
    >ere
    gtaagtgtacacgttatttctaatgtgaaacttttgacgacaaaagacatttatag

Based on these three alignments, we can conclude that all sites are consistent
across different alignment strategies because the three methods produced the
exact same alignment.

#### ConsAlign alignment

This is the summarized alignment produced by ConsAlign. The first sequence
represents the status of each site after comparing between the different
strategies. Here, consistent sites are marked by "C" and inconsistent sites
are marked by "N". Because all sites are consistent, all characters in the
marker sequence are "C".

    >marker
    CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC
    >mel01
    GTAAGTGTACACATTATTTCCGATGTGGGCCTTTTGACGACAAAAGAAATTTATAG
    >mel02
    GTAAGTGTACACATTATTTCCGATGTGGGCCTTTTGACGACAAAAGAAATTTATAG
    >sim
    GTAAGTGTACACATTATTTCGGATGTGGGTCTTTTGACGAC-AAAGACATTTATAG
    >yak
    GTATGTGTACACGTTATTTCTAATGTGAAACTTTTAACGAC-GAAGACATTTCTAG
    >ere
    GTAAGTGTACACGTTATTTCTAATGTGAAACTTTTGACGACAAAAGACATTTATAG

## Example - inconsistent sites

What happens if the alignment strategies produce different results?
The following example shows how alignments can change based on the
method used.

#### Unaligned

    >mel01
    gtaagatagtggcagattaattattagagtatctgcaacatgaatattatcttaacag
    >mel02
    gtaagatagtggcagattaattattagagtatctgcaacatgaatattatcttaacag
    >sim
    gtaagataatggcagattaaacattagattatctgcaacaagaatattatctcgacag
    >yak
    gtaagtctgtggcaggttaataattattataatatttgcaataacaatattttctgaacag
    >ere
    gtaagccagtggcaggttaataatcagtatatttgcaacaacaataattcctcaatag

#### Global alignment 

    >mel01
    gtaagatagtggcagat---taattattagagtatctgcaacatgaatattatcttaacag
    >mel02
    gtaagatagtggcagat---taattattagagtatctgcaacatgaatattatcttaacag
    >sim
    gtaagataatggcagat---taaacattagattatctgcaacaagaatattatctcgacag
    >yak
    gtaagtctgtggcaggttaataattattataatatttgcaataacaatattttctgaacag
    >ere
    gtaagccagtggcaggt---taataatcagtatatttgcaacaacaataattcctcaatag

#### Local alignment

    >mel01
    gtaagatagtggcagat---taattattagagtatctgcaacatgaatattatcttaacag
    >mel02
    gtaagatagtggcagat---taattattagagtatctgcaacatgaatattatcttaacag
    >sim
    gtaagataatggcagat---taaacattagattatctgcaacaagaatattatctcgacag
    >yak
    gtaagtctgtggcaggttaataattattataatatttgcaataacaatattttctgaacag
    >ere
    gtaagccagtggcaggt---taataatcagtatatttgcaacaacaataattcctcaatag

#### Affine-gap alignment

    >mel01
    gtaagatagtggcagattaattattaga---gtatctgcaacatgaatattatcttaacag
    >mel02
    gtaagatagtggcagattaattattaga---gtatctgcaacatgaatattatcttaacag
    >sim
    gtaagataatggcagattaaacattaga---ttatctgcaacaagaatattatctcgacag
    >yak
    gtaagtctgtggcaggttaataattattataatatttgcaataacaatattttctgaacag
    >ere
    gtaagccagtggcaggttaataatcagt---atatttgcaacaacaataattcctcaatag

In this case, the affine-gap method produced a different alignment compared
to the global and local alignments. Thus we observe that gap moves around the
middle portion of the alignment while the rest of the alignment remains the
same.

#### ConsAlign alignment

    >marker
    CCCCCCCCCCCCCCCCCNNNNNNNNNNNNNNCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC
    >mel01
    gtaagatagtggcagattaattattaga---gtatctgcaacatgaatattatcttaacag
    >mel02
    gtaagatagtggcagattaattattaga---gtatctgcaacatgaatattatcttaacag
    >sim
    gtaagataatggcagattaaacattaga---ttatctgcaacaagaatattatctcgacag
    >yak
    gtaagtctgtggcaggttaataattattataatatttgcaataacaatattttctgaacag
    >ere
    gtaagccagtggcaggttaataatcagt---atatttgcaacaacaataattcctcaatag

ConsAlign detects these inconsistencies between alignments and marks these
inconsistent sites with the character "N" in the marker sequence.

## Installation

### Requirements

- MAFFT must be installed in your system. ConsAlign does not include MAFFT.
- MAFFT must be included in the $PATH variable and executable from the shell.
  This means that the program can be started in any directory by simplying
  typing `mafft` from the command line.

ConsAlign is available as a compiled binary for Mac and Linux operating
systems.

- [Mac][2] - Compiled and tested on MacOS 10.11.6
- [Linux][3] - compiled and tested on Ubuntu Linux

For other systems, its possible to compile the program from source
using `go build`. Note that you must have [Go][4] installed in your system
to compile this program.

## Links

- [MAFFT download page][1]
- [ConsAlign for MacOS][2]
- [ConsAlign for Linux (compiled on Ubuntu 16.04 x86-64)][3]

[1]: http://mafft.cbrc.jp/alignment/software/
[2]: https://github.com/kentwait/consalign/releases/download/v1.0.1/consalign
[3]: https://github.com/kentwait/consalign/releases/download/v1.0.1/consalign_linux_amd64
