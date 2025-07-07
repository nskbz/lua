package ast

/*
chunk ::=block

block ::= {stat} [retstat]
retstat ::= return [explist] [‘;’]
explist ::= exp {‘,’ exp}
*/
type Block struct {
	LastLine int //用于debug
	Stats    []Stat
	RetExps  []Exp //没有返回语句，则为nil
}

type Chunk *Block
