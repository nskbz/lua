package ast

/*
chunk ::=block

block ::= {stat} [retstat]
retstat ::= return [explist] [‘;’]
explist ::= exp {‘,’ exp}
*/
type Block struct {
	LastLine int //?for what
	Stats    []Stat
	RetExps  []Exp
}

type Chunk *Block
