package hclparser

import (
	"github.com/fatih/color"
)

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

func (p *PrintParams) Shift() {
	p.Indent++
}

func (p *PrintParams) Unshift() {
	p.Indent--
	//Do not allow negative indent
	if p.Indent < 0 {
		p.Indent = 0
	}
}

func (p *PrintParams) GetIndentation() string {
	var ins string = ""
	for i := 0; i < p.Indent; i++ {
		ins = ins + p.indentUnit
	}
	return ins
}

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
