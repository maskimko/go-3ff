package hclparser

import (
	"3ff/diff"
	"3ff/utils"
	"fmt"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"log"
	"os"
	"strconv"
	"strings"
)

type Block hclsyntax.Block
type Blocks hclsyntax.Blocks
type Body hclsyntax.Body
type Attribute hclsyntax.Attribute
type Attributes struct {
	Mapped hclsyntax.Attributes
	List   []Attribute
}
type Expression struct {
	Contained hclsyntax.Expression
}
type Item struct {
	Contained hclsyntax.ObjectConsItem
}
type Expressions struct {
	List []Expression
}
type Items struct {
	List []Item
}
type SortableFiles []SortableFile
type SortableFile struct {
	File *os.File
}

func ConvertExpressionsHcl2HclS(in []hcl.Expression) []hclsyntax.Expression {
	elist := make([]hclsyntax.Expression, len(in), len(in))
	for i, v := range in {
		elist[i] = v.(hclsyntax.Expression)
	}
	return elist
}

func NewHclSyntaxExpressions(expressions []hclsyntax.Expression) *Expressions {
	elist := make([]Expression, len(expressions))
	i := 0
	for _, v := range expressions {
		elist[i] = Expression{Contained: v}
		i++
	}
	return &Expressions{List: elist}
}
func NewHclExpressions(expressions []hcl.Expression) *Expressions {
	return NewHclSyntaxExpressions(ConvertExpressionsHcl2HclS(expressions))
}
func NewItems(items []hclsyntax.ObjectConsItem) *Items {
	l := make([]Item, len(items))
	for i, v := range items {
		l[i] = Item{Contained: v}
	}
	return &Items{List: l}
}
func NewAttributes(attrs hclsyntax.Attributes) *Attributes {
	a := Attributes{Mapped: attrs}
	a.List = make([]Attribute, len(attrs))
	i := 0
	for _, v := range attrs {
		a.List[i] = Attribute(*v)
		i++
	}
	return &a
}
func (b Block) DiffParam() string {
	return fmt.Sprintf("%s.%s", b.Type, strings.Join(b.Labels, "."))
}

//func (b Block) DiffRepresentation() string {
//	return "Not implemented yet"
//}

func (a Attribute) DiffParam() string {
	return a.Name
}

//func (a Attribute) DiffRepresentation() string {
//	return "Not implemented yet"
//}

func (a Attribute) Range() hcl.Range {
	return a.NameRange
}

func (b Blocks) Len() int {
	return len(b)
}

func (b Blocks) Less(i, j int) bool {
	var bi Block = Block(*b[i])
	var bj Block = Block(*b[j])
	return bi.DiffParam() < bj.DiffParam()
}

func (b Blocks) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (a Attributes) Less(i, j int) bool {
	return a.List[i].Name < a.List[j].Name
}
func (a Attributes) Swap(i, j int) {
	a.List[i], a.List[j] = a.List[j], a.List[i]
}
func (a Attributes) Len() int {
	return len(a.Mapped)
}

func (e *Expressions) GetDiffables() *[]diff.Diffable {
	d := make([]diff.Diffable, len(e.List))
	i := 0
	for _, v := range e.List {
		d[i] = &Expression{Contained: v.Contained}
		i++
	}
	return &d
}

func (items *Items) GetDiffables() *[]diff.Diffable {
	d := make([]diff.Diffable, len(items.List))
	i := 0
	for _, v := range items.List {
		d[i] = &Item{Contained: v.Contained}
		i++
	}
	return &d
}

func (e *Expressions) Get(i int) *Expression {
	return &e.List[i]
}
func (items *Items) Get(i int) *Item {
	return &items.List[i]
}
func (i *Item) DiffParam() string {
	return (&Expression{Contained: i.Contained.KeyExpr}).DiffParam()
}
func (e *Expression) DiffParam() string {
	if te, ok := e.Contained.(*hclsyntax.TemplateExpr); ok {
		var tvals []string
		for _, pv := range te.Parts {
			tvals = append(tvals, (&Expression{Contained: pv}).DiffParam())
		}
		//Perhaps there should be adifferent format
		return strings.Join(tvals, ",")
	}
	if oce, ok := e.Contained.(*hclsyntax.ObjectConsExpr); ok {
		if len(oce.Items) > 0 {
			//Prefer name over id
			for _, item := range oce.Items {
				key := (&Expression{Contained: item.KeyExpr}).DiffParam()
				if key == "name" {
					val := (&Expression{Contained: item.ValueExpr}).DiffParam()
					return val
				}
			}
			for _, item := range oce.Items {
				key := (&Expression{Contained: item.KeyExpr}).DiffParam()
				if key == "id" {
					val := (&Expression{Contained: item.ValueExpr}).DiffParam()
					return val
				}
			}
			//Pick first value if no success
			return (&Expression{Contained: oce.Items[0].ValueExpr}).DiffParam()
		} else {
			return "[]"
		}
	}
	if we, ok := e.Contained.(*hclsyntax.FunctionCallExpr); ok {
		return we.Name
	}
	v, diag := e.Value(nil)
	if diag != nil && diag.HasErrors() {
		r := e.Range()
		val, err := utils.GetStringFromRange(r)
		if err != nil {
			if Debug {
				log.Printf("Failed to get value of expression. Diagnostic: %s File: %s (%d:%d-%d:%d) Error: %s",
					diag.Error(), r.Filename, r.Start.Line, r.Start.Column, r.End.Line, r.End.Column, err)
			}
			return "unavailable"
		}
		return val
	}

	if v.IsNull() {
		return "nil"
	}
	t := v.Type()
	if t.IsObjectType() {
		vm := v.AsValueMap()
		name, ok := vm["name"]
		if ok {
			return name.AsString()
		}
		id, ok := vm["id"]
		if ok {
			return id.AsString()
		}
	}
	if t.IsPrimitiveType() {
		switch t.FriendlyName() {
		case "bool":
			return strconv.FormatBool(v.True())
		case "number":
			return v.AsBigFloat().String()
		case "string":
			return v.AsString()
		}
	}
	//TODO: Implement other types
	return v.GoString()
}

func (e *Expression) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	return e.Contained.Value(ctx)
}
func (i *Item) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	return i.Contained.ValueExpr.Value(ctx)
}
func (e *Expression) Variables() []hcl.Traversal {
	return e.Contained.Variables()
}

func (e *Expression) Range() hcl.Range {
	return e.Contained.Range()
}

func (e *Expression) StartRange() hcl.Range {
	return e.Contained.StartRange()
}

func (e *Expressions) Less(i, j int) bool {
	idp := e.List[i].DiffParam()
	jdp := e.List[j].DiffParam()
	return idp < jdp
}
func (e *Expressions) Swap(i, j int) {
	e.List[i], e.List[j] = e.List[j], e.List[i]
}
func (e *Expressions) Len() int {
	return len(e.List)
}
func (items *Items) Less(i, j int) bool {
	idp := items.List[i].DiffParam()
	jdp := items.List[j].DiffParam()
	return idp < jdp
}
func (items *Items) Swap(i, j int) {
	items.List[i], items.List[j] = items.List[j], items.List[i]
}
func (items *Items) Len() int {
	return len(items.List)
}
func (a Attributes) GetDiffables() *[]diff.Diffable {
	d := make([]diff.Diffable, len(a.List))
	for i, v := range a.List {
		d[i] = v
	}
	return &d
}

func (b Blocks) Get(i int) *Block {
	var j Block = Block(*b[i])
	return &j
}
func (b Blocks) Range() hcl.Range {
	h := hclsyntax.Blocks(b)
	return h.Range()
}

func (b Blocks) GetDiffables() *[]diff.Diffable {
	d := make([]diff.Diffable, b.Len())
	for i, _ := range b {
		d[i] = b.Get(i)
	}
	return &d
}
func (b Block) Range() hcl.Range {
	h := hclsyntax.Block(b)
	return h.Range()
}

func (b Block) GetBody() *Body {
	h := hclsyntax.Block(b)
	body := Body(*h.Body)
	return &body
}

func (b Body) GetBlocks() Blocks {
	return Blocks(b.Blocks)
}

func (sf SortableFiles) Less(i, j int) bool {

	ifi, err := sf[i].File.Stat()
	if err != nil {
		log.Fatalf("Cannot perform stat() on File %s while sorting", sf[i].File.Name())
	}
	jfi, err := sf[j].File.Stat()
	if err != nil {
		log.Fatalf("Cannot perform stat() on File %s while sorting", sf[j].File.Name())
	}
	return ifi.Name() < jfi.Name()
}
func (sf SortableFiles) Swap(i, j int) {
	sf[i], sf[j] = sf[j], sf[i]
}
func (sf SortableFiles) Len() int {
	return len(sf)
}
func (sf SortableFile) DiffParam() string {
	return sf.File.Name()
}

//func (sf SortableFile) DiffRepresentation() string {
//	return "Not implemented yet"
//}
