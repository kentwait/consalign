package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ExecMafft calls the MAFFT program with the given arguments
func ExecMafft(mafftCmd string, args []string) string {
	absPath, lookErr := exec.LookPath(mafftCmd)
	if lookErr != nil {
		panic(lookErr)
	}

	cmd := exec.Command(absPath, args...)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(stdout))
		return ""
	}
	return string(stdout)
}

// EinsiAlign calls MAFFT to align sequences by local alignment with
// affine-gap scoring.
func EinsiAlign(mafftCmd, fastaPath string, iterations int) (stdout string) {
	args := []string{
		"--quiet",
		"--genafpair",
		"--maxiterate",
		strconv.Itoa(iterations),
		fastaPath,
	}
	stdout = ExecMafft(mafftCmd, args)
	return
}

// LinsiAlign calls MAFFT to align sequences by local alignment.
func LinsiAlign(mafftCmd, fastaPath string, iterations int) (stdout string) {
	args := []string{
		"--quiet",
		"--localpair",
		"--maxiterate",
		strconv.Itoa(iterations),
		fastaPath,
	}
	stdout = ExecMafft(mafftCmd, args)
	return
}

// GinsiAlign calls MAFFT to align sequences by global alignment.
func GinsiAlign(mafftCmd, fastaPath string, iterations int) (stdout string) {
	args := []string{
		"--quiet",
		"--globalpair",
		"--maxiterate",
		strconv.Itoa(iterations),
		fastaPath,
	}
	stdout = ExecMafft(mafftCmd, args)
	return
}

// Sequence is an interface for single character sequences stored as a string
// and multi-character sequences stored as a slice.
type Sequence interface {
	UngappedCoords(string) []int
	UngappedPositionSlice(string) []int
	ToUpper()
	ToLower()
}

// CharSequence is a struct for nucleotide and single-letter protein sequences.
type CharSequence struct {
	id    string
	title string
	seq   string
}

// UngappedCoords returns the positions in the sequence where the character
// does not match the gap character.
func (s *CharSequence) UngappedCoords(gapChar string) (colCoords []int) {
	set := make(map[int]struct{})
	for j := 0; j < len(s.seq); j++ {
		if string(s.seq[j]) != gapChar {
			set[j] = struct{}{}
		}
	}
	for key := range set {
		colCoords = append(colCoords, key)
	}
	sort.Ints(colCoords)
	return
}

// UngappedPositionSlice returns a slice that counts only over characters
// that does not match the gap character in the sequence.
// If a character matches the gap character, -1 is inserted instead of the
// ungapped count.
func (s *CharSequence) UngappedPositionSlice(gapChar string) (arr []int) {
	cnt := 0
	for j := 0; j < len(s.seq); j++ {
		if string(s.seq[j]) != gapChar {
			arr = append(arr, cnt)
			cnt++
		} else {
			arr = append(arr, -1)
		}
	}
	return
}

// ToUpper changes the case of the sequence to all uppercase letters.
func (s *CharSequence) ToUpper() {
	s.seq = strings.ToUpper(s.seq)
}

// ToLower changes the case of the sequence to all lowercase letters.
func (s *CharSequence) ToLower() {
	s.seq = strings.ToLower(s.seq)
}

// SequenceAlignment is a slice of Sequence pointers.
type SequenceAlignment []*CharSequence

// UngappedCoords returns the row and column positions in the sequence alignment
// where the character does not match the gap character.
func (a SequenceAlignment) UngappedCoords(gapChar string) (rowCoords, colCoords []int) {
	var currColCoords []int
	for i, s := range a {
		// s := reflect.ValueOf(sPtr)
		currColCoords = s.UngappedCoords(gapChar)
		for c := 0; c < len(currColCoords); c++ {
			rowCoords = append(rowCoords, i)
		}
		colCoords = append(colCoords, currColCoords...)
	}
	return
}

// UngappedPositionMatrix returns a matrix that counts only over characters
// that does not match the gap character for each sequence in the alignment.
// If a character in a sequence matches the gap character, -1 is inserted
// instead of the ungapped count.
func (a SequenceAlignment) UngappedPositionMatrix(gapChar string) (m [][]int) {
	for _, s := range a {
		m = append(m, s.UngappedPositionSlice(gapChar))
	}
	return
}

// ToUpper changes the case of all sequences to all uppercase letters.
func (a SequenceAlignment) ToUpper() {
	for _, s := range a {
		s.seq = strings.ToUpper(s.seq)
	}
}

// ToLower changes the case of of all sequences to all lowercase letters.
func (a SequenceAlignment) ToLower() {
	for _, s := range a {
		s.seq = strings.ToLower(s.seq)
	}
}

// ToFastaString returns the FASTA-formatted string of the sequence alignment.
func (a SequenceAlignment) ToFastaString() string {
	return SequencesToString(a)
}

// ToFasta saves the sequence alignment to a FASTA file.
func (a SequenceAlignment) ToFasta(path string) {
	WriteBufferToFile(path, SequencesToBuffer(a))
}

// StringToCharSequences loads a string generated from properly formatted
// FASTA file into a SequenceAlignment struct.
func StringToCharSequences(s string) (sequences SequenceAlignment) {
	lines := strings.Split(s, "\n")

	var id, title string
	var seqBuffer bytes.Buffer
	var splitted []string

	for _, line := range lines {
		if strings.HasPrefix(line, ">") {
			if seqBuffer.Len() > 0 {
				sequences = append(sequences, &CharSequence{id, title, seqBuffer.String()})
				seqBuffer.Reset()
			}
			splitted = strings.SplitN(line[1:], " ", 2)
			id = splitted[0]
			if len(splitted) == 2 {
				title = splitted[1]
			}
		} else if strings.HasPrefix(line, "\n") {
			continue
		} else if strings.HasPrefix(line, "#") {
			continue
		} else {
			seqBuffer.WriteString(line)
		}
	}
	if seqBuffer.Len() > 0 {
		sequences = append(sequences, &CharSequence{id, title, seqBuffer.String()})
	}
	return
}

// SequencesToBuffer converts sequences in the sequene alignment into a buffered
// stream which can then be converted to bytes or a string.
func SequencesToBuffer(a SequenceAlignment) bytes.Buffer {
	var buffer bytes.Buffer
	// Append each Sequence in SequenceAlignment
	for _, s := range a {
		if len(s.title) > 0 {
			buffer.WriteString(fmt.Sprintf(">%s %s\n", s.id, s.title))
		} else {
			buffer.WriteString(fmt.Sprintf(">%s\n", s.id))
		}
		buffer.WriteString(s.seq + "\n")
	}
	return buffer
}

// SequencesToString converts sequences in the sequene alignment into a FASTA
// formatted string.
func SequencesToString(a SequenceAlignment) string {
	buffer := SequencesToBuffer(a)
	return buffer.String()
}

// ConsistentAlignmentPositions returns the list of positions in the alignment
// that are considered consistent given by the alignment pattern per site
// across all given alignments.
func ConsistentAlignmentPositions(gapChar string, matrices ...[][]int) []bool {
	// Assumes all matrices have the same shape
	m := len(matrices[0])
	n := len(matrices[0][0])

	pos := make([]bool, n)
	for j := 0; j < n; j++ {
		// For the current column, compare the column pattern across the
		// matrices. If a difference in corresponding values are detected,
		// then the pattern is deemed inconsistent and the current column
		// is therefore also inconsistent.
		consistent := true

	PatternLoop:
		for k := 1; k < len(matrices); k++ {
			for i := 0; i < m; i++ {
				if matrices[0][i][j] != matrices[k][i][j] {
					consistent = false
					break PatternLoop
				}
			}
		}

		if consistent == true {
			pos[j] = true
		} else {
			pos[j] = false
		}
	}
	return pos
}

// BufferedMarkedAlignment writes a marked multiple sequence alignment
// in the FASTA format to the buffer.
func BufferedMarkedAlignment(template SequenceAlignment, consistentPos []bool, markerID, consistentMarker, inconsistentMarker string) bytes.Buffer {
	var buffer bytes.Buffer

	// Append marker sequence
	buffer.WriteString(fmt.Sprintf(">%s\n", markerID))
	for _, t := range consistentPos {
		if t == true {
			buffer.WriteString(consistentMarker)
		} else {
			buffer.WriteString(inconsistentMarker)
		}
	}
	buffer.WriteString("\n")

	// Append each Sequence in SequenceAlignment
	for _, s := range template {
		if len(s.title) > 0 {
			buffer.WriteString(fmt.Sprintf(">%s %s\n", s.id, s.title))
		} else {
			buffer.WriteString(fmt.Sprintf(">%s\n", s.id))
		}
		buffer.WriteString(s.seq + "\n")
	}
	return buffer
}

// ConsistentAlignmentPipeline aligns using global, local, and affine-local alignment
// strategies to determine positions that have a consistent alignment pattern over
// the three different strategies.
func ConsistentAlignmentPipeline(inputPath, gapChar, markerID, consistentMarker, inconsistentMarker string, iterations int, toUpper, toLower, saveTempAlns bool) bytes.Buffer {

	const mafftCmd = "mafft"

	ginsiAln := StringToCharSequences(GinsiAlign(mafftCmd, inputPath, iterations))
	linsiAln := StringToCharSequences(LinsiAlign(mafftCmd, inputPath, iterations))
	einsiAln := StringToCharSequences(EinsiAlign(mafftCmd, inputPath, iterations))

	if saveTempAlns == true {
		ginsiAln.ToFasta(inputPath + ".ginsi.aln")
		linsiAln.ToFasta(inputPath + ".linsi.aln")
		einsiAln.ToFasta(inputPath + ".einsi.aln")
	}

	consistentPos := ConsistentAlignmentPositions(
		"-",
		ginsiAln.UngappedPositionMatrix(gapChar),
		linsiAln.UngappedPositionMatrix(gapChar),
		einsiAln.UngappedPositionMatrix(gapChar),
	)

	if toUpper == true {
		einsiAln.ToUpper()
	} else if toLower == true {
		einsiAln.ToLower()
	}

	return BufferedMarkedAlignment(einsiAln, consistentPos, markerID, consistentMarker, inconsistentMarker)
}

// WriteBufferToFile writes the contents of a buffer to file.
func WriteBufferToFile(path string, b bytes.Buffer) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b.WriteTo(f)
	f.Sync()
}

// exists returns whether the given file or directory exists or not,
// and accompanying errors.
func exists(path string) (bool, error) {
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

	maxIterPtr := flag.Int("maxiterate", 0, "Maximum number of iterative refinement that MAFFT will perform. (default: 0)")
	gapCharPtr := flag.String("gapchar", "-", "Character in the alignment used to represent a gap. (default: -)")
	markerIDPtr := flag.String("marker_id", "marker", "Name of marker sequence. (default: marker)")
	cMarkerPtr := flag.String("consistent_marker", "C", "Character to indicate a site is consistent across all alignment strategies. (default: C)")
	icMarkerPtr := flag.String("inconsistent_marker", "N", "Character to indicate a site is inconsistent in at least one alignment strategy. (default: N)")
	changeCasePtr := flag.String("change_case", "upper", "Change the case of the sequences. {upper|lower|no}")
	saveTempAlnPtr := flag.Bool("save_temp_alignments", false, "Save G-INSI, L-INSI, E-INSI alignments generated by MAFFT.")
	isBatchPtr := flag.String("batch", "", "Run in batch mode reading all the files in the folder.")
	inSuffixPtr := flag.String("input_suffix", ".fa", "Only files ending with this suffix will be processed. Used in conjunction with -batch.")
	outSuffixPtr := flag.String("output_suffix", ".aln", "Suffix to be appended to the end of the filename of resulting alignments. Used in conjunction with -batch.")
	outDirPtr := flag.String("outdir", "", "Output directory where alignments will be saved. Used in conjunction with -batch.")

	flag.Parse()

	if len(*isBatchPtr) < 1 {
		// Single file mode
		args := flag.Args()
		switch {
		case len(args) < 1:
			os.Stderr.WriteString("Error: Missing path to FASTA file.\n")
			os.Exit(1)
		case len(args) > 1:
			os.Stderr.WriteString("Error: More than 1 positional argument passed.\n")
			os.Exit(1)
		case len(args) == 1:
			if doesExist, _ := exists(args[0]); doesExist == false {
				os.Stderr.WriteString("Error: file does not exist.\n")
				os.Exit(1)
			}
		}

		switch {
		case *changeCasePtr == "lower":
			toLower = true
		case *changeCasePtr == "upper":
			toUpper = true
		}

		buffer := ConsistentAlignmentPipeline(args[0], *gapCharPtr, *markerIDPtr, *cMarkerPtr, *icMarkerPtr, *maxIterPtr, toUpper, toLower, *saveTempAlnPtr)

		fmt.Print(buffer.String())
	} else {
		// Batch mode
		if doesExist, _ := exists(*isBatchPtr); doesExist == false {
			os.Stderr.WriteString("Error: Specified directory containing FASTA files does not exist.\n")
			os.Exit(1)
		}

		// Check if outdir flag used
		// Check if folder exists
		switch {
		case len(*outDirPtr) < 1:
			os.Stderr.WriteString("Error: Missing output directory.\nUse -outdir to specify an output directory where alignments will be saved.\n")
			os.Exit(1)
		case len(*outDirPtr) == 1:
			if doesExist, _ := exists(*outDirPtr); doesExist == false {
				os.Stderr.WriteString("Error: Specified output directory does not exist.\n")
				os.Exit(1)
			}
		}

		// Read all fasta files in directory matching suffix
		files, err := filepath.Glob(*isBatchPtr + "/*" + *inSuffixPtr)
		if err != nil {
			panic(err)
		}
		var outputPath string
		for _, f := range files {
			buffer := ConsistentAlignmentPipeline(f, *gapCharPtr, *markerIDPtr, *cMarkerPtr, *icMarkerPtr, *maxIterPtr, toUpper, toLower, *saveTempAlnPtr)
			outputPath = *outDirPtr + "/" + filepath.Base(f) + *outSuffixPtr
			WriteBufferToFile(outputPath, buffer)
		}
	}

}