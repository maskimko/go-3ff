//hclparser package contains functions and classes to work with HCL data sctructures
//as with entities which can be compared in order to compute diff
package hclparser

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/maskimko/go-3ff/diff"
	"github.com/maskimko/go-3ff/utils"
	"log"
	"reflect"
	"sort"
	"strings"
)

var TerraformOutput bool

//ChangedExprContext holds the details of the change
type ChangedExprContext struct {
	//If modificationType is greater than 0 it means that attribute was added
	//If it is less than 0 it - attribute was removed
	//If it equals 0 - attribute value was changed
	//Unchanged attributes should not apppear in this structure
	ModificationType            int8
	Orig, Modified              hclsyntax.Expression
	OrigDiffValue, ModifDiffVal string
}

//ExpressionDiff struct holds the diff data between HCL data objects
type ExpressionDiff struct {
	//ExpressionDiff can nest smaller Nested ones
	Nested *ExpressionDiff
	//Changed flag shows the status of HCL data object
	Changed bool
	//Changes is an array of ChangedExprContext of this ExpressionDiff
	Changes []ChangedExprContext
}

//ChangedAttributeContext struct holds the attribute change details
type ChangedAttributeContext struct {
	//If modificationType is greater than 0 it means that attribute was added
	//If it is less than 0 it - attribute was removed
	//If it equals 0 - attribute value was changed
	//Unchanged attributes should not apppear in this structure
	ModificationType int8
	Orig, Modif      *Attribute
	Diff             *ExpressionDiff
}
type AttributesDiff struct {
	Changes []ChangedAttributeContext
}

func (atdf *AttributesDiff) HasChanges() bool {
	if atdf.Changes == nil || len(atdf.Changes) == 0 {
		return false
	} else {
		return true
	}
}

func (edf *ExpressionDiff) HasChanges() bool {
	if edf.Changes == nil || len(edf.Changes) == 0 {
		return false
	} else {
		return true
	}
}
func (edf *ExpressionDiff) Add(ctx ChangedExprContext) {
	edf.Changes = append(edf.Changes, ctx)
}

func (atdf *AttributesDiff) Add(ctx ChangedAttributeContext) {
	atdf.Changes = append(atdf.Changes, ctx)
}

//analyzeAttributesDiff function computes the difference of HCL Attributes by given path and returns AttributesDiff
func (mr *ModifiedResources) analyzeAttributesDiff(orig, modif *Attributes, path []string) *AttributesDiff {
	attributesDiff := &AttributesDiff{Changes: make([]ChangedAttributeContext, 0)}
	if modif.Len() == 0 && orig.Len() == 0 {
		return attributesDiff
	}

	sort.Sort(modif)
	sort.Sort(orig)
	o := *orig.GetDiffables()
	m := *modif.GetDiffables()
	_, subset := diff.GetLongestCommonSubsequence(&o, &m)
	subs := *subset
	for i, j, k := 0, 0, 0; j < orig.Len() || i < modif.Len(); {
		if j < len(o) {
			if k < len(subs) && subs[k].DiffParam() == o[j].DiffParam() {
				if subs[k].DiffParam() == m[i].DiffParam() {
					oa := orig.Mapped[o[j].DiffParam()]
					ma := modif.Mapped[m[i].DiffParam()]
					edchan := make(chan *ExpressionDiff)
					go asyncExpressionDiff(oa.Expr, ma.Expr, edchan)
					ed := <-edchan
					if ed.Changed {

						attributesDiff.Add(ChangedAttributeContext{Orig: &orig.List[j], Modif: &modif.List[i], ModificationType: 0, Diff: ed})
					}
					i++
					j++
					k++
				} else {
					attributesDiff.Add(ChangedAttributeContext{Modif: &modif.List[i], ModificationType: 1})
					i++
				}
			} else {
				attributesDiff.Add(ChangedAttributeContext{Orig: &orig.List[j], ModificationType: -1})
				j++
			}
		} else {
			attributesDiff.Add(ChangedAttributeContext{Modif: &modif.List[i], ModificationType: 1})
			i++
		}
	}
	if attributesDiff.HasChanges() {
		mr.Add(strings.Join(path, "/"))
	}
	return attributesDiff
}

//analyzeBlocksDiff function computes the difference of HCL Blocks by given path and prints out it using PrintParams formatting
func (mr *ModifiedResources) analyzeBlocksDiff(orig, modif Blocks, path []string, p *PrintParams) bool {

	result := true
	sort.Sort(orig)
	sort.Sort(modif)
	o := *orig.GetDiffables()
	m := *modif.GetDiffables()
	_, lcs := diff.GetLongestCommonSubsequence(&o, &m)
	subs := *lcs
	for i, j, k := 0, 0, 0; j < len(o) || i < len(m); {
		if j < len(o) {
			if k < len(subs) && subs[k].DiffParam() == o[j].DiffParam() {
				if subs[k].DiffParam() == m[i].DiffParam() {
					result = mr.computeBlockDiff(orig.Get(j), modif.Get(i), path) && result
					i++
					j++
					k++
				} else {
					PrintRemoved(m[i], p)
					mr.Add(m[i].DiffParam())
					result = false
					i++
				}
			} else {
				PrintAdded(o[j], p)
				mr.Add(o[j].DiffParam())
				result = false
				j++
			}
		} else {
			PrintRemoved(m[i], p)
			mr.Add(m[i].DiffParam())
			result = false
			i++
		}
	}
	return result
}

//PrintRemoved function prints out removed diff.Diffable using PrintParams formatting
func PrintRemoved(mr diff.Diffable, p *PrintParams) {
	if !TerraformOutput {
		fmt.Printf("%s%s %s\n", p.GetIndentation(), p.RemoveColor.Sprint("-"), mr.DiffParam())
	}
}

//PrintAdded function prints out added diff.Diffable using PrintParams formatting
func PrintAdded(mr diff.Diffable, p *PrintParams) {
	if !TerraformOutput {
		fmt.Printf("%s%s %s\n", p.GetIndentation(), p.AddColor.Sprintf("+"), mr.DiffParam())
	}
}

//PrintModified function prints modified string using PrintParams formatting
func PrintModified(name string, p *PrintParams) {
	if !TerraformOutput {
		fmt.Printf("%s%s %s\n", p.GetIndentation(), p.ChangedColor.Sprintf("~"), name)
	}
}

//printRemovedAttribute function prints out removed attribute using PrintParams formatting
func printRemovedAttribute(name string, p *PrintParams) {
	if !TerraformOutput {
		fmt.Printf("%s%s [%s]\n", p.GetIndentation(), p.RemoveColor.Sprint("-"), name)
	}
}

//printAddedAttribute function prints out added attribute using PrintParams formatting
func printAddedAttribute(name string, p *PrintParams) {
	if !TerraformOutput {
		fmt.Printf("%s%s [%s]\n", p.GetIndentation(), p.ChangedColor.Sprintf("~"), name)
	}
}

//printModifiedAttributeWithDeepDiff prints out modifications with a recursive indentation using PrintParams formatting
func printModifiedAttributeWithDeepDiff(modification ChangedAttributeContext, p *PrintParams) {
	if !TerraformOutput {
		name := modification.Orig.Name
		if name != modification.Modif.Name {
			if Debug {
				log.Println("WARNING!!!Attribute name changed. In this method this should never happen!!!")
			}
		}
		fmt.Printf("%s%s [%s]  %s\n", p.GetIndentation(), p.ChangedColor.Sprintf("~"), name, color.MagentaString("("))
		printExpressionDiff(p.GetIndentation()+strings.Repeat(" ", len(name)+4)+"   ", modification.Diff, p)
		fmt.Printf("%s%s\n", p.GetIndentation()+strings.Repeat(" ", len(name)+4)+"  ", color.MagentaString(")"))
	}
}

//printExpressionDiff function prints the Expression diff using PrintParams for formatting
func printExpressionDiff(indent string, ed *ExpressionDiff, p *PrintParams) {
	if ed == nil {
		return
	}
	if !ed.Changed {
		if Debug {
			log.Println("WARNING!!! Print of not changed expression diff. This should never happen")
		}
		return
	}
	for _, v := range ed.Changes {
		printExpressionContext(indent, v, p)
	}
	printExpressionDiff(indent+"\t", ed.Nested, p)

}

//printExpressionContext function prints the Expression diff with a context using PrintParams for formatting
func printExpressionContext(indent string, cec ChangedExprContext, p *PrintParams) {

	if cec.ModificationType < 0 {
		r := cec.Orig.Range()
		fmt.Printf("%s%s '%s'\t%s\n", indent, p.AddColor.Sprintf("+"), cec.OrigDiffValue, formatRange(r, p))
	} else if cec.ModificationType > 0 {
		r := cec.Modified.Range()
		fmt.Printf("%s%s '%s'\t%s\n", indent, p.RemoveColor.Sprintf("-"), cec.ModifDiffVal, formatRange(r, p))
	} else {
		fmt.Printf("%s%s '%s' %s '%s'\t%s%s%s\n", indent, p.ChangedColor.Sprintf("~"), cec.OrigDiffValue,
			p.ChangedColor.Sprintf("->"), cec.ModifDiffVal, formatRange(cec.Orig.Range(), p), p.ChangedColor.Sprintf("->"), formatRange(cec.Modified.Range(), p))
	}
}

func formatRange(r hcl.Range, p *PrintParams) string {
	return p.LocationColor.Sprintf("@%s[%d:%d-%d:%d]", r.Filename, r.Start.Line, r.Start.Column, r.End.Line, r.End.Column)
}

//PrintAttributeContext function prints out AttributesDiff using PrintParams formatting
func PrintAttributeContext(atdf *AttributesDiff, p *PrintParams) {
	for _, v := range atdf.Changes {
		if v.ModificationType == 0 {
			printModifiedAttributeWithDeepDiff(v, p)
		} else if v.ModificationType > 0 {
			printAddedAttribute(v.Modif.Name, p)
		} else {
			printRemovedAttribute(v.Orig.Name, p)
		}
	}
}

//computeHclSyntaxExpressionsDiff function is implementation of Myers diff algorithm for HCL Expression.
//This function computes difference and stores values to the ExpressionDiff
func computeHclSyntaxExpressionsDiff(otce, mtce []hclsyntax.Expression) *ExpressionDiff {
	ed := ExpressionDiff{Changes: make([]ChangedExprContext, 0), Changed: false}
	otces := NewHclSyntaxExpressions(otce)
	sort.Sort(otces)
	mtces := NewHclSyntaxExpressions(mtce)
	sort.Sort(mtces)

	//TODO: Try to remove this
	od := otces.GetDiffables()
	md := mtces.GetDiffables()

	_, s := diff.GetLongestCommonSubsequence(od, md)
	subs := *s
	for i, j, k := 0, 0, 0; j < otces.Len() || i < mtces.Len(); {
		if j < otces.Len() {
			if k < len(subs) && subs[k].DiffParam() == otces.Get(j).DiffParam() {
				if subs[k].DiffParam() == mtces.Get(i).DiffParam() {
					var edchan chan *ExpressionDiff = make(chan *ExpressionDiff)
					go asyncExpressionDiff(otces.Get(j).Contained, mtces.Get(i).Contained, edchan)
					ied := <-edchan
					if ied.Changed {
						ed.Add(ChangedExprContext{Orig: otces.Get(j).Contained, Modified: mtces.Get(i).Contained,
							ModificationType: 0, OrigDiffValue: otces.Get(j).DiffParam(), ModifDiffVal: mtces.Get(i).DiffParam()})
						ed.Nested = ied
						ed.Changed = true
					}
					i++
					j++
					k++
				} else {
					ed.Add(ChangedExprContext{Modified: mtces.Get(i).Contained, ModificationType: 1, ModifDiffVal: mtces.Get(i).DiffParam()})
					ed.Changed = true
					i++
				}
			} else {
				ed.Add(ChangedExprContext{Orig: otces.Get(j).Contained, ModificationType: -1, OrigDiffValue: otces.Get(j).DiffParam()})
				ed.Changed = true
				j++
			}
		} else {
			ed.Add(ChangedExprContext{Modified: mtces.Get(i).Contained, ModificationType: 1, ModifDiffVal: mtces.Get(i).DiffParam()})
			ed.Changed = true
			i++
		}
	}
	return &ed
}

//computeItemsDiff fucntion is implementaion of Myers diff algorithm for HCL Items
func computeItemsDiff(o, m []hclsyntax.ObjectConsItem) *ExpressionDiff {
	ed := ExpressionDiff{Changes: make([]ChangedExprContext, 0), Changed: false}
	origItems := NewItems(o)
	sort.Sort(origItems)
	modifiedItems := NewItems(m)
	sort.Sort(modifiedItems)
	_, s := diff.GetLongestCommonSubsequence(origItems.GetDiffables(), modifiedItems.GetDiffables())
	subs := *s
	for i, j, k := 0, 0, 0; j < origItems.Len() || i < modifiedItems.Len(); {
		if j < origItems.Len() {
			if k < len(subs) && subs[k].DiffParam() == origItems.Get(j).DiffParam() {
				if subs[k].DiffParam() == modifiedItems.Get(i).DiffParam() {
					ied := analyzeExpressionDiff(origItems.Get(j).Contained.ValueExpr, modifiedItems.Get(i).Contained.ValueExpr)
					if ied.Changed {
						ed.Add(ChangedExprContext{Orig: origItems.Get(j).Contained.ValueExpr, Modified: modifiedItems.Get(i).Contained.ValueExpr,
							ModificationType: 0, OrigDiffValue: origItems.Get(j).DiffParam(), ModifDiffVal: modifiedItems.Get(i).DiffParam()})
						ed.Nested = ied
						ed.Changed = true
					}
					i++
					j++
					k++
				} else {
					ed.Add(ChangedExprContext{Modified: modifiedItems.Get(i).Contained.ValueExpr, ModificationType: 1, ModifDiffVal: modifiedItems.Get(i).DiffParam()})
					ed.Changed = true
					i++
				}
			} else {
				ed.Add(ChangedExprContext{Orig: origItems.Get(j).Contained.ValueExpr, ModificationType: -1, OrigDiffValue: origItems.Get(j).DiffParam()})
				ed.Changed = true
				j++
			}
		} else {
			ed.Add(ChangedExprContext{Modified: modifiedItems.Get(i).Contained.ValueExpr, ModificationType: 1, ModifDiffVal: modifiedItems.Get(i).DiffParam()})
			ed.Changed = true
			i++
		}
	}
	return &ed
}

//asyncExpressionDiff function does the same as analyzeExpressionDiff in asynchronous way and puts values to the diff channel
func asyncExpressionDiff(orig, modif hclsyntax.Expression, diff chan *ExpressionDiff) {
	ed := analyzeExpressionDiff(orig, modif)
	diff <- ed
	close(diff)
}

//analyzeExpressionDiff function is implementation of Myers diff algorithm to compute the diff of two HCL expressions
func analyzeExpressionDiff(orig, modif hclsyntax.Expression) *ExpressionDiff {
	if orig == nil && modif == nil {
		return &ExpressionDiff{Changes: make([]ChangedExprContext, 0)}
	}
	if orig == nil {

		return &ExpressionDiff{Changes: []ChangedExprContext{{Modified: modif, ModificationType: 1,
			ModifDiffVal: utils.GetStringFromHclSyntaxExpression(modif)}}, Changed: true}
	}
	if modif == nil {
		return &ExpressionDiff{Changes: []ChangedExprContext{{Orig: orig, ModificationType: -1,
			OrigDiffValue: utils.GetStringFromHclSyntaxExpression(orig)}}, Changed: true}
	}

	if otce, ok := orig.(*hclsyntax.TupleConsExpr); ok {
		//Case the same
		if mtce, ok := modif.(*hclsyntax.TupleConsExpr); ok {
			return computeHclSyntaxExpressionsDiff(otce.Exprs, mtce.Exprs)
		} else {
			//If type mismatch than the whole expression is a difference
			changes := []ChangedExprContext{{Orig: orig, Modified: modif, ModificationType: 0,
				OrigDiffValue: utils.GetStringFromHclSyntaxExpression(orig), ModifDiffVal: utils.GetStringFromHclSyntaxExpression(modif)}}
			return &ExpressionDiff{Changes: changes, Changed: true}
		}
	}
	if otmple, ok := orig.(*hclsyntax.TemplateExpr); ok {
		if mtmple, ok := modif.(*hclsyntax.TemplateExpr); ok {
			return computeHclSyntaxExpressionsDiff(otmple.Parts, mtmple.Parts)
		} else {
			//If type mismatch than the whole expression is a difference
			changes := []ChangedExprContext{{Orig: orig, Modified: modif, ModificationType: 0,
				OrigDiffValue: utils.GetStringFromHclSyntaxExpression(orig), ModifDiffVal: utils.GetStringFromHclSyntaxExpression(modif)}}
			return &ExpressionDiff{Changes: changes, Changed: true}
		}
	}
	//Sort diffable and so on!!!
	if oxtme, ok := orig.(*hclsyntax.ObjectConsExpr); ok {
		if mxtme, ok := modif.(*hclsyntax.ObjectConsExpr); ok {
			return computeItemsDiff(oxtme.Items, mxtme.Items)
		} else {
			// If type has changed, than all expression change
			return makeFullExpressionDiff(orig, modif)
		}

	}
	if _, ok := orig.(*hclsyntax.LiteralValueExpr); ok {
		if _, ok := modif.(*hclsyntax.LiteralValueExpr); ok {
			oVal := utils.GetStringFromHclSyntaxExpression(orig)
			mVal := utils.GetStringFromHclSyntaxExpression(modif)
			if oVal != mVal {
				return makeFullExpressionDiff(orig, modif)
			} else {
				return &ExpressionDiff{Changes: make([]ChangedExprContext, 0), Changed: false}
			}
		} else {
			return makeFullExpressionDiff(orig, modif)
		}
	}
	if otwe, ok := orig.(*hclsyntax.TemplateWrapExpr); ok {
		if mtwe, ok := modif.(*hclsyntax.TemplateWrapExpr); ok {
			var edchan chan *ExpressionDiff = make(chan *ExpressionDiff)
			go asyncExpressionDiff(otwe.Wrapped, mtwe.Wrapped, edchan)
			return <-edchan
			//return analyzeExpressionDiff(otwe.Wrapped, mtwe.Wrapped)
		} else {
			return makeFullExpressionDiff(orig, modif)
		}
	}
	if ofce, ok := orig.(*hclsyntax.FunctionCallExpr); ok {
		if mfce, ok := modif.(*hclsyntax.FunctionCallExpr); ok {
			if ofce.Name == mfce.Name {
				ed := &ExpressionDiff{Changes: make([]ChangedExprContext, 0), Changed: false}
				ned := computeHclSyntaxExpressionsDiff(ofce.Args, mfce.Args)
				if ned.Changed {
					ed.Changed = true
					ed.Nested = ned
				}
				return ed
			} else {
				return makeFullExpressionDiff(orig, modif)
			}
		} else {
			return makeFullExpressionDiff(orig, modif)
		}
	}
	if _, ok := orig.(*hclsyntax.ScopeTraversalExpr); ok {
		//if mfce, ok := modif.(*hclsyntax.ScopeTraversalExpr); ok {
		//
		//} else {
		return makeFullExpressionDiff(orig, modif)
		//}
	}
	if _, ok := orig.(*hclsyntax.SplatExpr); ok {
		return makeFullExpressionDiff(orig, modif)
	}
	if _, ok := orig.(*hclsyntax.BinaryOpExpr); ok {
		return makeFullExpressionDiff(orig, modif)
	}
	if _, ok := orig.(*hclsyntax.IndexExpr); ok {
		return makeFullExpressionDiff(orig, modif)
	}
	if _, ok := orig.(*hclsyntax.RelativeTraversalExpr); ok {
		return makeFullExpressionDiff(orig, modif)
	}
	if _, ok := orig.(*hclsyntax.ForExpr); ok {
		return makeFullExpressionDiff(orig, modif)
	}
	if _, ok := orig.(*hclsyntax.UnaryOpExpr); ok {
		return makeFullExpressionDiff(orig, modif)
	}
	log.Printf("WARNING!!! Unknown expression type %s", reflect.TypeOf(orig))
	return makeFullExpressionDiff(orig, modif)
}

//makeFullExpressionDiff function computes diff of two Expressions and stores it in ExpressionDiff
func makeFullExpressionDiff(orig, modif hclsyntax.Expression) *ExpressionDiff {
	ed := &ExpressionDiff{Changes: make([]ChangedExprContext, 0), Changed: false}
	origVal := utils.GetStringFromHclSyntaxExpression(orig)
	modifVal := utils.GetStringFromHclSyntaxExpression(modif)
	ed.Changed = origVal != modifVal
	ed.Add(ChangedExprContext{Orig: orig, Modified: modif, ModificationType: 0, OrigDiffValue: origVal, ModifDiffVal: modifVal})
	return ed
}
