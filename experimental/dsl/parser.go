package main

import (
	"fmt"
	"log"
)

func parseView(strs []string) view {
	view := view{}
	for _, str := range strs {
		expr, err := parseExpression(str)
		if err != nil {
			log.Printf("%s: %v", str, err)
			continue
		}
		log.Printf("%s: OK", str)
		view = append(view, expr)
	}
	return view
}

func parseExpression(str string) (expression, error) {
	var (
		expr expression
		not  bool
	)
	_, c := lex(str)
	for item := range c {
		switch item.itemType {
		case itemNot:
			not = !not

		case itemAll:
			expr.selector = selectAll

		case itemConnected:
			expr.selector = selectConnected

		case itemTouched:
			expr.selector = selectTouched

		case itemLike:
			item = <-c
			switch item.itemType {
			case itemRegex:
				expr.selector = selectLike(item.literal)
			default:
				return expression{}, fmt.Errorf("bad WITH: want %s, got %s", itemRegex, item.itemType)
			}

		case itemWith:
			item = <-c
			switch item.itemType {
			case itemKeyValue:
				expr.selector = selectWith(item.literal)
			default:
				return expression{}, fmt.Errorf("bad WITH: want %s, got %s", itemKeyValue, item.itemType)
			}

		case itemRemove:
			expr.transformer = transformRemove

		case itemShowOnly:
			expr.transformer = transformShowOnly

		case itemMerge:
			expr.transformer = transformMerge

		case itemGroupBy:
			item = <-c
			switch item.itemType {
			case itemKeyList:
				expr.transformer = transformGroupBy(item.literal)
			default:
				return expression{}, fmt.Errorf("bad WITH: want %s, got %s", itemKeyList, item.itemType)
			}

		default:
			return expression{}, fmt.Errorf("%s: %s", str, item.literal)
		}
	}
	if not {
		expr.selector = selectNot(expr.selector)
	}
	return expr, nil
}
