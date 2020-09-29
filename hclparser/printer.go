package hclparser

import (
	"github.com/fatih/color"
)

//PrintParams class holds the knowledge of formatting the output
type PrintParams struct {
	RemoveColor   *color.Color
	AddColor      *color.Color
	ChangedColor  *color.Color
	OkColor       *color.Color
	CommentColor  *color.Color
	LocationColor *color.Color
	Indent        int
	indentUnit    string
}

//Shift function increases indentation of the PrintParams
func (p *PrintParams) Shift() {
	p.Indent++
}

//Unshift function decreases indentation of the PrintParams
func (p *PrintParams) Unshift() {
	p.Indent--
	//Do not allow negative indent
	if p.Indent < 0 {
		p.Indent = 0
	}
}

//GetIndentation function returns the indentation size value
func (p *PrintParams) GetIndentation() string {
	var ins string = ""
	for i := 0; i < p.Indent; i++ {
		ins = ins + p.indentUnit
	}
	return ins
}

//GetDefaultPrintParams return the default formatting style
func GetDefaultPrintParams() *PrintParams {
	pp := PrintParams{
		RemoveColor:   color.New(color.FgRed),
		AddColor:      color.New(color.FgGreen),
		ChangedColor:  color.New(color.FgYellow),
		CommentColor:  color.New(color.FgMagenta),
		OkColor:       color.New(color.FgWhite),
		LocationColor: color.New(color.FgHiBlack + color.Underline),
		indentUnit:    "\t"}
	return &pp
}
