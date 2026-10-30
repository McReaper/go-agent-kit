// Package proof runs every tool the kit names on the trial project and
// checks each one reports exactly what the trial's cases expect: nothing on
// trial/clean, and each case's annotated findings on its violation.
//
// A case is a directory under trial/cases holding the files it adds to (or
// replaces in) a copy of trial/clean, and a want.txt listing its findings,
// one per line:
//
//	<tool> <file>:<line> <regexp matching the message>
//
// Lines starting with # are comments; "!isolated" runs the case alone
// rather than with the others; "!delete <file>" removes a clean file first.
package proof
