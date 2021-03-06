// parse parenthetical Lisp notation into syntax tree

package parse

import (
	"fmt"
	"regexp"
)

// Find shortest sequence of double quote, followed by escaped and unescaped characters, followed by double quote
var stringRegex = regexp.MustCompile(`"([^"\\]|\\.)*"`)

// Find a single quoted character, which consists of a single quote, an optional backslash, any character, and a closing single quote.
var charRegex = regexp.MustCompile(`'\\?.'`)

// Find longest sequence of non-syntax characters
var tokenRegex = regexp.MustCompile("[^ \t\n\"()]+")

// Given a string starting with a token, find the end of the token (the index of the first character that follows the token).
// Rules:
// * If the token starts with a double quote ("), the token ends at the next double quote that isn't backslash-escaped.
// * Otherwise the token ends right before the next syntax character.
// Returns a negative value if there is no valid token.
func findTokenEnd(s string) int {
	var re *regexp.Regexp
	switch s[0] {
	case '"':
		re = stringRegex
	case '\'':
		re = charRegex
	default:
		re = tokenRegex
	}
	loc := re.FindStringIndex(s)
	if loc == nil {
		return -1
	} else {
		return loc[1] // ending index of regex match
	}
}

// change the node according to how the indentation depth level changed
// NOTE: This assumes that the calling parse function is not at a blank line state.
func indentSrfi49(depthChange int, node *Node) *Node {
	switch {
	case depthChange < 0:
		// decrease depth by as many levels as the change
		for i := depthChange; i < 0; i++ {
			node = node.Parent()
		}
		// Now that we're back at the right level, we need to start a new sibling Node.
		node = node.Parent().MakeChild()
	case depthChange == 0:
		// make new sibling node at same depth
		node = node.Parent().MakeChild()
	case depthChange > 0:
		// increasing depth makes a new child node
		for i := 0; i < depthChange; i++ {
			node = node.MakeChild()
		}
	default:
		panic("Impossible! DepthChange is not less than, equal to, or greater than zero!")
	}
	return node
}

// Parse a Golid string into its syntax tree, automatically
// determining which top-level blocks of code use which syntax.
func parseString(s string) (Expression, error) {
	root := Root() // top-level node

	// process a top-level node
	doTopNode := func() error {
		isens := true // Indentation SENSitivity
		if s[0] == '(' {
			isens = false
			s = s[1:] // don't make the same node twice
		}
		tabDepth := 0 // for indent-grouping syntax
		n := root.MakeChild()
	loop: // go until break or out of code
		for s != "" {
			switch s[0] {
			case '(': // go deeper
				n = n.MakeChild()
				s = s[1:]
			case ')': // go up
				n = n.Parent()
				s = s[1:]
			case ' ', '\t': // ignore mid-line whitespace
				s = s[1:]
			case '\n': // whitespace, but may end node
				// get to the good stuff
				for s != "" && s[0] == '\n' {
					s = s[1:]
				}
				if !isens { // paren-only syntax
					if n == root {
						break loop
					} else {
						continue loop
					}
				}
				if s == "" || s[0] != '\t' {
					break loop
				}
				newDepth := 0
				for s != "" && s[0] == '\t' {
					newDepth++
					s = s[1:]
				}
				n = indentSrfi49(newDepth-tabDepth, n)
				tabDepth = newDepth
			case ';': // skip rest of line
				for s != "" && s[0] != '\n' {
					s = s[1:]
				}
				s = s[1:] // alse skip trailing '\n'
			default: // must be a token, finally
				end := findTokenEnd(s)
				if end < 0 {
					return fmt.Errorf("Could not find end of token %s.", s)
				} else {
					n.AddToken(s[0:end])
					s = s[end:]
				}
			}
		}
		return nil
	}

	// process the entire string
	for s != "" {
		// get to a top-level block
		switch s[0] {
		case ' ', '\t': // whitespace to skip
			// TODO: handle " (" case of indent-grouped syntax starting with
			// a paren group. I think it's the only case where unified
			// parsing doesn't automatically detect correctly between
			// classic and indent Lisp syntaxes.
			s = s[1:]
		case '\n': // newline to note and pass
			s = s[1:]
		case ';': // comment to skip
			for s[0] != '\n' {
				s = s[1:]
			}
		default: // we've reached the next node
			err := doTopNode()
			if err != nil {
				return nil, err
			}
		}
	}

	// remove extra layers (e.g., if this was ran for less than a file)
	for root != nil && root.first == root.last && root.content == "" {
		root = root.first
	}

	return root, nil
}
