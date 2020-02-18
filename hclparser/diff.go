package hclparser

import (
	"3ff/utils"
	"fmt"
	"github.com/fatih/color"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"log"
	"reflect"

	"3ff/diff"
	"sort"
	"strings"
)

var TerraformOutput bool

type ChangedExprContext struct {
	//If modificationType is greater than 0 it means that attribute was added
	//If it is less than 0 it - attribute was removed
	//If it equals 0 - attribute value was changed
	//Unchanged attributes should not apppear in this structure
	ModificationType            int8
	Orig, Modified              hclsyntax.Expression
	OrigDiffValue, ModifDiffVal string
}
type ExpressionDiff struct {
	Nested  *ExpressionDiff
	Changed bool
	Changes []ChangedExprContext
}

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
func (mr *ModifiedResources) analyzeAttributesDiff(orig, modif *Attributes, path []string, p *PrintParams) *AttributesDiff {
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
					var edchan chan *ExpressionDiff = make(chan *ExpressionDiff)
					go asyncExpressionDiff(oa.Expr, ma.Expr, edchan)
					//ed := analyzeExpressionDiff(oa.Expr, ma.Expr)
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

func (mr *ModifiedResources) analyzeBlocksDiff(orig, modif Blocks, path []string, p *PrintParams) bool {

	result := true
	sort.Sort(orig)
	sort.Sort(modif)
	o := *orig.GetDiffables()
	m := *modif.GetDiffables()
	_, lcs := diff.GetLongestCommonSubsequence(&o, &m)
	//log.Println("Found longest common subsequence:", n)
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

func PrintRemoved(mr diff.Diffable, p *PrintParams) {
	if !TerraformOutput {
		fmt.Printf("%s%s %s\n", p.GetIndentation(), p.RemoveColor.Sprint("-"), mr.DiffParam())
	}
}

func PrintAdded(mr diff.Diffable, p *PrintParams) {
	if !TerraformOutput {
		fmt.Printf("%s%s %s\n", p.GetIndentation(), p.AddColor.Sprintf("+"), mr.DiffParam())
	}
}

func PrintModified(name string, p *PrintParams) {
	if !TerraformOutput {
		fmt.Printf("%s%s %s\n", p.GetIndentation(), p.ChangedColor.Sprintf("~"), name)
	}
}

func printRemovedAttribute(name string, p *PrintParams) {
	if !TerraformOutput {
		fmt.Printf("%s%s [%s]\n", p.GetIndentation(), p.RemoveColor.Sprint("-"), name)
	}
}

func printAddedAttribute(name string, p *PrintParams) {
	if !TerraformOutput {
		fmt.Printf("%s%s [%s]\n", p.GetIndentation(), p.ChangedColor.Sprintf("~"), name)
	}
}

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

func printExpressionContext(indent string, cec ChangedExprContext, p *PrintParams) {

	if cec.ModificationType < 0 {
		r := cec.Orig.Range()
		fmt.Printf("%s%s '%s'\t%s\n", indent, p.AddColor.Sprintf("+"), cec.OrigDiffValue, formatRange(r, p))
	} else if cec.ModificationType > 0 {
		r := cec.Modified.Range()
		fmt.Printf("%s%s '%s'\t%s\n", indent, p.RemoveColor.Sprintf("-"), cec.ModifDiffVal, formatRange(r, p))
	} else {
		//fmt.Printf("%s%s %s %s %s\t%s %s\n", indent, p.ChangedColor.Sprintf("~"), cec.OrigDiffValue,
		//	color.YellowString("->"), cec.ModifDiffVal,formatRange(or,p),formatRange(mr,p))
		fmt.Printf("%s%s '%s' %s '%s'\t%s%s%s\n", indent, p.ChangedColor.Sprintf("~"), cec.OrigDiffValue,
			p.ChangedColor.Sprintf("->"), cec.ModifDiffVal, formatRange(cec.Orig.Range(), p), p.ChangedColor.Sprintf("->"), formatRange(cec.Modified.Range(), p))
	}
}

func formatRange(r hcl.Range, p *PrintParams) string {
	return p.LocationColor.Sprintf("@%s[%d:%d-%d:%d]", r.Filename, r.Start.Line, r.Start.Column, r.End.Line, r.End.Column)
}

func PrintAttributeContext(atdf *AttributesDiff, p *PrintParams) {
	for _, v := range atdf.Changes {
		if v.ModificationType == 0 {
			//var logString string = ""
			//var err error
			//logString, err = utils.GetChangeLogString(v.Orig.Expr.Range(), v.Modif.Expr.Range())
			//if err != nil && Debug {
			//	log.Print("Cannot compose attribute diff")
			//}
			//if Debug {
			//	log.Printf("Attribute was changed\n"+
			//		"Original: %s (in File %s from line %d, column %d till line %d, column %d)\n"+
			//		"Modified: %s (in File %s from line %d, column %d till line %d, column %d)",
			//		v.Orig.Name, v.Orig.SrcRange.Filename, v.Orig.Expr.Range().Start.Line, v.Orig.Expr.Range().Start.Column, v.Orig.Expr.Range().End.Line, v.Orig.Expr.Range().End.Column,
			//		v.Modif.Name, v.Modif.SrcRange.Filename, v.Modif.Expr.Range().Start.Line, v.Modif.Expr.Range().Start.Column, v.Modif.Expr.Range().End.Line, v.Modif.Expr.Range().End.Column)
			//
			//}
			//if logString != "" {
			//printModifiedAttributeWithDiff(v.Orig.Name, logString, p)
			printModifiedAttributeWithDeepDiff(v, p)
			//} else {
			//	printModifiedAttribute(v.Orig.Name, p)
			//}
		} else if v.ModificationType > 0 {
			//if Debug {
			//	log.Printf("Attribute was added\n"+
			//		"Modified: %s (in File %s from line %d, column %d till line %d, column %d)\n",
			//		v.Modif.Name, v.Modif.SrcRange.Filename, v.Modif.Expr.Range().Start.Line, v.Modif.Expr.Range().Start.Column, v.Modif.Expr.Range().End.Line, v.Modif.Expr.Range().End.Column)
			//}
			printAddedAttribute(v.Modif.Name, p)
		} else {
			//if Debug {
			//	log.Printf("Attribute was removed\n"+
			//		"Original: %s (in File %s from line %d, column %d till line %d, column %d)\n",
			//		v.Orig.Name, v.Orig.SrcRange.Filename, v.Orig.Expr.Range().Start.Line, v.Orig.Expr.Range().Start.Column, v.Orig.Expr.Range().End.Line, v.Orig.Expr.Range().End.Column)
			//}
			printRemovedAttribute(v.Orig.Name, p)
		}
	}
}

func computeHclExpressionsDiff(orig, modif []hcl.Expression) *ExpressionDiff {
	return computeHclSyntaxExpressionsDiff(ConvertExpressionsHcl2HclS(orig), ConvertExpressionsHcl2HclS(modif))
}
func computeHclSyntaxExpressionsDiff(otce, mtce []hclsyntax.Expression) *ExpressionDiff {
	ed := ExpressionDiff{Changes: make([]ChangedExprContext, 0), Changed: false}
	otces := NewHclSyntaxExpressions(otce)
	sort.Sort(otces)
	//for m, xp := range otces.List {
	//	log.Printf("\tOriginal attribute expression %d: %s", m, xp)
	//}
	mtces := NewHclSyntaxExpressions(mtce)
	sort.Sort(mtces)
	//for m, xp := range mtces.List {
	//	log.Printf("\tModified attribute expressions %d: %s", m, xp)
	//}

	//TODO: Try to remove this
	od := otces.GetDiffables()
	md := mtces.GetDiffables()

	_, s := diff.GetLongestCommonSubsequence(od, md)
	subs := *s
	for i, j, k := 0, 0, 0; j < otces.Len() || i < mtces.Len(); {
		if j < otces.Len() {
			if k < len(subs) && subs[k].DiffParam() == otces.Get(j).DiffParam() {
				if subs[k].DiffParam() == mtces.Get(i).DiffParam() {
					//TODO: Use recursion here
					//r := expressionEquals(otces.Get(j).Contained, mtces.Get(i).Contained)
					var edchan chan *ExpressionDiff = make(chan *ExpressionDiff)
					go asyncExpressionDiff(otces.Get(j).Contained, mtces.Get(i).Contained, edchan)
					//ied := analyzeExpressionDiff(otces.Get(j).Contained, mtces.Get(i).Contained)
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

func asyncExpressionDiff(orig, modif hclsyntax.Expression, diff chan *ExpressionDiff) {
	ed := analyzeExpressionDiff(orig, modif)
	diff <- ed
	close(diff)
}

func analyzeExpressionDiff(orig, modif hclsyntax.Expression) *ExpressionDiff {
	//ed := &ExpressionDiff{Changes: make([]ChangedExprContext, 0), Changed: false}
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
			changes := []ChangedExprContext{ChangedExprContext{Orig: orig, Modified: modif, ModificationType: 0,
				OrigDiffValue: utils.GetStringFromHclSyntaxExpression(orig), ModifDiffVal: utils.GetStringFromHclSyntaxExpression(modif)}}
			return &ExpressionDiff{Changes: changes, Changed: true}
		}
	}
	if otmple, ok := orig.(*hclsyntax.TemplateExpr); ok {
		if mtmple, ok := modif.(*hclsyntax.TemplateExpr); ok {
			return computeHclSyntaxExpressionsDiff(otmple.Parts, mtmple.Parts)
		} else {
			//If type mismatch than the whole expression is a difference
			changes := []ChangedExprContext{ChangedExprContext{Orig: orig, Modified: modif, ModificationType: 0,
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

func makeFullExpressionDiff(orig, modif hclsyntax.Expression) *ExpressionDiff {
	ed := &ExpressionDiff{Changes: make([]ChangedExprContext, 0), Changed: false}
	origVal := utils.GetStringFromHclSyntaxExpression(orig)
	modifVal := utils.GetStringFromHclSyntaxExpression(modif)
	ed.Changed = origVal != modifVal
	ed.Add(ChangedExprContext{Orig: orig, Modified: modif, ModificationType: 0, OrigDiffValue: origVal, ModifDiffVal: modifVal})
	return ed
}
