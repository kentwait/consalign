package main

import (
	"fmt"
	"sort"
	"strings"
)

// Sequence is an interface for single character sequences stored as a string
// and multi-character sequences stored as a slice.
type Sequence interface {
	ID() string
	Title() string
	Sequence() string
	Char(int) string
	SetSequence(string)
	ToUpper()
	ToLower()
	UngappedCoords(string) []int
	UngappedPositionSlice(string) []int
}

// CharSequence struct for storing single-character biological sequences such
// as nucleotides and single-letter amino acids. However, any sequence that
// whose element can be represented as a single string character can be stored
// in CharSequence.
type CharSequence struct {
	id    string
	title string
	seq   string
}

// ID returns the id field of CharSequence.
func (s *CharSequence) ID() string {
	return s.id
}

// Title returns the title field of CharSequence.
func (s *CharSequence) Title() string {
	return s.title
}

// Sequence returns the seq field of CharSequence.
func (s *CharSequence) Sequence() string {
	return s.seq
}

// Char returns a single character from the seq field of CharSequence.
func (s *CharSequence) Char(i int) string {
	return string([]rune(s.seq)[i])
}

// SetSequence assigns a string to the seq field of CharSequence.
func (s *CharSequence) SetSequence(seq string) {
	s.seq = seq
}

// UngappedCoords returns the positions in the sequence where the character
// does not match the gap character.
func (s *CharSequence) UngappedCoords(gapChar string) (colCoords []int) {
	set := make(map[int]struct{})
	// Assumes gapChar contains only a "single character"
	// Convert single character string to rune slice, taking the first item
	gapRune := []rune(gapChar)[0]
	// Range over rune slice, j counts by Unicode code points, s is the rune representation of the character
	for j, s := range []rune(s.seq) {
		// If sequence rune is not a gap character rune, add to rune position to set, 0-indexed
		if s != gapRune {
			set[j] = struct{}{} // Uses empty anonymous struct
		}
	}
	// Range over set of positions
	// Since this is a map, order is scrambled
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
	// Assumes gapChar contains only a "single character"
	// Convert single character string to rune slice, taking the first item
	gapRune := []rune(gapChar)[0]
	cnt := 0
	for _, s := range []rune(s.seq) {
		// If sequence rune is not a gap character rune, append current count value to array and increment
		if s != gapRune {
			arr = append(arr, cnt)
			cnt++
			// If it is equal to the gap character rune, then append a -1.
			// Do not increment.
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

// CodonSequence is a struct for specifically designed for triplet nucleotide
// codon sequences. It embeds the CharSequence struct which also gives it
// id, title and seq fields. Additionally, CodonSequence has a prot field which
// stores a string and a codon string field which stores a slice of strings.
// The seq, prot and codons fields follow a positional correspondence.
// The first item in the codons slice translates to the first character
// in the prot string. The first item in the codons slice is equal to
// the first three characters of the seq string. This codon-seq correspondence
// should be consistent across the entire sequence.
type CodonSequence struct {
	CharSequence
	prot   string
	codons []string
}

// NewCodonSequence is a constructor that creates a new CodonSequence where
// prot and codons field values are automatically computed from the provided
// nucleotide sequence.
func NewCodonSequence(id, title, seq string) *CodonSequence {
	if len(seq)%3 != 0 {
		panic(fmt.Sprintf("Given seq's length (%d) not divisible by 3", len(seq)))
	}
	s := new(CodonSequence)
	s.id = id
	s.title = title
	s.SetSequence(seq)
	return s
}

// ID returns the id field of CodonSequence.
func (s *CodonSequence) ID() string {
	return s.id
}

// Title returns the title field of CodonSequence.
func (s *CodonSequence) Title() string {
	return s.title
}

// Sequence returns the seq field of CodonSequence. The seq field contains
// a nucleotide sequence stored as a string.
func (s *CodonSequence) Sequence() string {
	return s.seq
}

// Codons returns the codon field of CodonSequence. The codon field
// contains a nucleotide sequence delimited by codon. This is stored
// as a slice of 3-character strings.
func (s *CodonSequence) Codons() []string {
	return s.codons
}

// Prot returns the prot field of CodonSequence. The prot field
// contains the translated amino acid sequence based on the seq
// field using the standard genetic code. The amino acid sequence
// is encoded as single-character amino acids and stored as a
// string.
func (s *CodonSequence) Prot() string {
	return s.prot
}

// Char returns a single nucleotide from the seq field of CodonSequence.
func (s *CodonSequence) Char(i int) string {
	return string([]rune(s.seq)[i])
}

// ProtChar returns a single amino acid from the prot field of CodonSequence.
func (s *CodonSequence) ProtChar(i int) string {
	return string(s.prot[i])
}

// Codon returns a single codon 3 nucleotides long from the codons field of
// CodonSequence.
func (s *CodonSequence) Codon(i int) string {
	return string(s.codons[i])
}

/* The following two methods are setters for sequence fields in CodonSequence.
   Note that there is not method to set a protein sequence in the prot field.
   Because of the relationships between seq, prot, and codons, it is impossible
   to compute the values of seq and codons from the protein sequence alone.
   Although a protein sequence can be set literally, this is not recommended as
   there is no way to ensure that the relationships between seq, prot, and
   codons are maintained.
*/

// SetSequence assigns a nucleotide sequence to the seq field of CodonSequence.
// It also automatically fills the codons and prot fields by splitting the
// nucleotide sequence into triplets and translating each codon into its
// corresponding amino acid using the standard genetic code respectively.
func (s *CodonSequence) SetSequence(seq string) {
	// Converts sequence to rune slice to deal with unicode chars
	seqRune := []rune(seq)
	if len(seqRune)%3 != 0 {
		panic(fmt.Sprintf("Length of given seq \"%s\" is not divisible by 3", seq))
	}
	// Overwrite value of .seq
	s.seq = seq
	// Overwrites value of .codons
	var codons []string
	for i := 0; i < len(seqRune); i += 3 {
		codons = append(codons, string(seqRune[i:i+3]))
	}
	s.codons = codons
	// Overwrites the value of .prot
	s.prot = Translate(seq).String()
}

// SetCodons assigns a nucleotide sequence delimited by codon to the codons
// field of CodonSequence. It also automatically fills the seq and prot
// fields by joining the codons into a single continuous string and
// translating each codon into its corresponding amino acid using the
// standard genetic code respectively.
func (s *CodonSequence) SetCodons(seq []string) {
	s.codons = seq
	s.seq = strings.Join(seq, "")
	s.prot = Translate(s.seq).String()
}

// UngappedCoords returns the positions in the sequence where the character
// does not match the gap character.
func (s *CodonSequence) UngappedCoords(gapChar string) (colCoords []int) {
	if len(gapChar)%3 != 0 {
		panic(fmt.Sprintf("Length of given gapChar \"%s\" is not equal to 3", gapChar))
	}
	set := make(map[int]struct{})
	for j := 0; j < len(s.codons); j++ {
		if s.codons[j] != gapChar {
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
func (s *CodonSequence) UngappedPositionSlice(gapChar string) (arr []int) {
	if len(gapChar)%3 != 0 {
		panic(fmt.Sprintf("Length of given gapChar \"%s\" is not equal to 3", gapChar))
	}
	cnt := 0
	for j := 0; j < len(s.codons); j++ {
		if s.codons[j] != gapChar {
			arr = append(arr, cnt)
			cnt++
		} else {
			arr = append(arr, -1)
		}
	}
	return
}

// ToUpper changes the case of the sequence to all uppercase letters.
func (s *CodonSequence) ToUpper() {
	s.seq = strings.ToUpper(s.seq)
	s.prot = strings.ToUpper(s.prot)
	for i := 0; i < len(s.codons); i++ {
		s.codons[i] = strings.ToUpper(s.codons[i])
	}
}

// ToLower changes the case of the sequence to all lowercase letters.
func (s *CodonSequence) ToLower() {
	s.seq = strings.ToLower(s.seq)
	s.prot = strings.ToLower(s.prot)
	for i := 0; i < len(s.seq); i++ {
		s.codons[i] = strings.ToLower(s.codons[i])
	}
}

// sequence constants

var bases = [4]string{"T", "C", "A", "G"}
var codons = [64]string{
	"TTT", "TTC", "TTA", "TTG",
	"TCT", "TCC", "TCA", "TCG",
	"TAT", "TAC", "TAA", "TAG",
	"TGT", "TGC", "TGA", "TGG",
	"CTT", "CTC", "CTA", "CTG",
	"CCT", "CCC", "CCA", "CCG",
	"CAT", "CAC", "CAA", "CAG",
	"CGT", "CGC", "CGA", "CGG",
	"ATT", "ATC", "ATA", "ATG",
	"ACT", "ACC", "ACA", "ACG",
	"AAT", "AAC", "AAA", "AAG",
	"AGT", "AGC", "AGA", "AGG",
	"GTT", "GTC", "GTA", "GTG",
	"GCT", "GCC", "GCA", "GCG",
	"GAT", "GAC", "GAA", "GAG",
	"GGT", "GGC", "GGA", "GGG",
}
var stopCodons = [3]string{"TGA", "TAG", "TAA"}
var aminoAcids = [20]string{
	"A",
	"R",
	"N",
	"D",
	"C",
	"Q",
	"E",
	"G",
	"H",
	"I",
	"L",
	"K",
	"M",
	"F",
	"P",
	"S",
	"T",
	"W",
	"Y",
	"V",
}
var geneticCode = map[string]string{
	"TTT": "F",
	"TTC": "F",
	"TTA": "L",
	"TTG": "L",
	"TCT": "S",
	"TCC": "S",
	"TCA": "S",
	"TCG": "S",
	"TAT": "Y",
	"TAC": "Y",
	"TAA": "*",
	"TAG": "*",
	"TGT": "C",
	"TGC": "C",
	"TGA": "*",
	"TGG": "W",
	"CTT": "L",
	"CTC": "L",
	"CTA": "L",
	"CTG": "L",
	"CCT": "P",
	"CCC": "P",
	"CCA": "P",
	"CCG": "P",
	"CAT": "H",
	"CAC": "H",
	"CAA": "Q",
	"CAG": "Q",
	"CGT": "R",
	"CGC": "R",
	"CGA": "R",
	"CGG": "R",
	"ATT": "I",
	"ATC": "I",
	"ATA": "I",
	"ATG": "M",
	"ACT": "T",
	"ACC": "T",
	"ACA": "T",
	"ACG": "T",
	"AAT": "N",
	"AAC": "N",
	"AAA": "K",
	"AAG": "K",
	"AGT": "S",
	"AGC": "S",
	"AGA": "R",
	"AGG": "R",
	"GTT": "V",
	"GTC": "V",
	"GTA": "V",
	"GTG": "V",
	"GCT": "A",
	"GCC": "A",
	"GCA": "A",
	"GCG": "A",
	"GAT": "D",
	"GAC": "D",
	"GAA": "E",
	"GAG": "E",
	"GGT": "G",
	"GGC": "G",
	"GGA": "G",
	"GGG": "G",
	"---": "-",
}
