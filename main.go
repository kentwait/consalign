package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Exists returns whether the given file or directory Exists or not,
// and accompanying errors.
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func main() {
	toUpper := false
	toLower := false

	// # Program arguments
	// ConsPos uses flag arguments to set run parameters.
	// There are 4 types of flags based on what parameter they set.
	// - ConsPos flags set general parameters regarding the ConsPos program itself.
	// - Codon-specific flags set parameters for when dealing with codon alignments.
	// - Batch flags indicate that the analysis is a batch analysis of many multiple sequence alignments. Arguments under this category sets file input and output handling for files.
	// - MAFFT-related flags are arguments intended for the MAFFT alignment program that ConsPos uses to generate the multiple sequence alignments.

	// ConsPos flags
	markerIDPtr := flag.String("marker_id", "marker", "Name of marker sequence.")
	cMarkerPtr := flag.String("consistent_marker", "C", "Character to indicate a site is consistent across all alignment strategies.")
	icMarkerPtr := flag.String("inconsistent_marker", "N", "Character to indicate a site is inconsistent in at least one alignment strategy.")
	gapCharPtr := flag.String("gapchar", "-", "Character in the alignment used to represent a gap.")
	changeCasePtr := flag.String("change_case", "upper", "Change the case of the sequences. {upper|lower|no}")

	// Codon-specific flags
	isCodonPtr := flag.Bool("codon", false, "Create a codon-based alignment.")

	// Batch flags
	isBatchPtr := flag.String("batch", "", "Run in batch mode which reads files found in the specified folder.")
	outDirPtr := flag.String("outdir", "", "Output directory where alignments will be saved. Used in conjunction with -batch.")
	inSuffixPtr := flag.String("input_suffix", ".fa", "Only files ending with this suffix will be processed. Used in conjunction with -batch.")
	outSuffixPtr := flag.String("output_suffix", ".aln", "Suffix to be appended to the end of the filename of resulting alignments. Used in conjunction with -batch.")

	// MAFFT-related flags
	maxIterPtr := flag.Int("maxiterate", 1, "Maximum number of iterative refinement that MAFFT will perform.")
	saveTempAlnPtr := flag.Bool("save_temp_alignments", false, "Save G-INSI, L-INSI, E-INSI alignments generated by MAFFT.")
	mafftPathPtr := flag.String("mafft_path", "mafft", "Path to MAFFT executable. If MAFFT is registered in $PATH, you can use \"mafft\".")

	flag.Parse()

	// Checks if values of arguments are valid.

	// Validates supplied path for MAFFT executable.
	// Raises an error and exits if the path does not exist.
	if _, lookErr := exec.LookPath(*mafftPathPtr); lookErr != nil {
		os.Stderr.WriteString("Error: Invalid MAFFT path. Make sure that the MAFFT executable is installed and is accessible at the path specified in -mafft_path.\n")
		os.Exit(1)
	}

	// The program is two modes: single file and batch mode.
	// Because arguments are mode-dependent, the validity of arguments are checked depending whether or not -batch is empty (single file) or not (batch mode).
	if len(*isBatchPtr) == 0 {
		// Single file mode expects a single positional argument (FASTA file path).
		// Checks whether there is at least one positional argument present.
		// Raises an error and exists if no positional arguments are present, or when more than one is given.
		args := flag.Args()
		if len(args) == 0 {
			os.Stderr.WriteString("Error: Missing path to FASTA file.\n")
			os.Exit(1)
		} else if len(args) > 1 {
			os.Stderr.WriteString("Error: More than 1 positional argument passed.\n")
			os.Exit(1)
		}
		// Given that there is only one positional argument supplied, checks whether a file exists at that path.
		// This does not check whether the file is a FASTA file though.
		if doesExist, _ := Exists(args[0]); doesExist == false {
			os.Stderr.WriteString("Error: file does not exist.\n")
			os.Exit(1)
		}

		// Converts case change choices to boolean variables.
		switch *changeCasePtr {
		case "lower":
			toLower = true
		case "upper":
			toUpper = true
		case "no":
		default:
			os.Stderr.WriteString("Error: Invalid -change_case value {upper|lower|no}.\n")
			os.Exit(1)
		}

		// The program further splits into two more modes depending on whether the sequences should be treated as single character sites or codons (3 characters per site) and call the appropriate function.
		// The gapchar argument depends on this.
		// For example, if codons, the gapchar should be 3 characters long, and only a single character if not.
		var buffer bytes.Buffer
		if *isCodonPtr {
			// TODO: gapchar check should be length, not char matching
			if *gapCharPtr == "-" {
				*gapCharPtr = "---"
			}
			buffer = ConsistentCodonAlnPipeline(args[0], *gapCharPtr, *markerIDPtr, *cMarkerPtr, *icMarkerPtr, *maxIterPtr, toUpper, toLower, *saveTempAlnPtr)
		} else {
			buffer = ConsistentAlnPipeline(args[0], *gapCharPtr, *markerIDPtr, *cMarkerPtr, *icMarkerPtr, *maxIterPtr, toUpper, toLower, *saveTempAlnPtr)
		}
		fmt.Print(buffer.String())
		// TODO: clear buffer after writing to stdout?

	} else {
		// Batch mode is activated when the value of -batch is not empty.
		// The -batch value is the directory containing the FASTA files to work on.
		// Checks whether the path exists but it does not check if the path is a file or a directory.
		if doesExist, _ := Exists(*isBatchPtr); doesExist == false {
			os.Stderr.WriteString("Error: Specified directory containing FASTA files does not exist.\n")
			os.Exit(1)
		}

		// In batch mode, -outdir must be specified to tell ConsPos where to save the aligned files.
		// Check is -outdir has any value.
		if len(*outDirPtr) == 0 {
			os.Stderr.WriteString("Error: Missing output directory.\nUse -outdir to specify an output directory where alignments will be saved.\n")
			os.Exit(1)
		}
		// Checks if the specified output directory path exists.
		// This does not check if the path is to a directory or a file.
		// This does not check if the path (assuming its a directory) is empty.
		if doesExist, _ := Exists(*outDirPtr); doesExist == false {
			os.Stderr.WriteString("Error: Specified output directory does not exist.\n")
			os.Exit(1)
		}

		// Read all fasta files in directory matching suffix
		files, err := filepath.Glob(*isBatchPtr + "/*" + *inSuffixPtr)
		if err != nil {
			panic(err)
		}

		// Check whether to treat sequences as codon alignments or not and call the appropriate function
		var outputPath string
		var buffer bytes.Buffer
		for _, f := range files {
			if *isCodonPtr {
				buffer = ConsistentCodonAlnPipeline(f, *gapCharPtr, *markerIDPtr, *cMarkerPtr, *icMarkerPtr, *maxIterPtr, toUpper, toLower, *saveTempAlnPtr)
			} else {
				buffer = ConsistentAlnPipeline(f, *gapCharPtr, *markerIDPtr, *cMarkerPtr, *icMarkerPtr, *maxIterPtr, toUpper, toLower, *saveTempAlnPtr)
			}
			outputPath = *outDirPtr + "/" + filepath.Base(f) + *outSuffixPtr
			WriteBufferToFile(outputPath, buffer)

			buffer.Reset()
		}
	}
}
