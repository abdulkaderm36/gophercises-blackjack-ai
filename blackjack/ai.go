package blackjack

import (
	"fmt"

	deck "github.com/abdulkaderm36/gophercises-deck"
)

type AI interface {
	Bet(shuffled bool) int
	Play(hand []deck.Card, dealer deck.Card) Move
	Results(hands [][]deck.Card, dealer []deck.Card)
}

type dealerAI struct{}

func (ai dealerAI) Bet(shuffled bool) int {
	// noop
	return 1
}

func (ai dealerAI) Play(hand []deck.Card, dealer deck.Card) Move {
	dScore := Score(hand...)
	if dScore <= 16 || (dScore == 17 && Soft(hand...)) {
		return MoveHit
	}
	return MoveStand
}

func (ai dealerAI) Results(hands [][]deck.Card, dealer []deck.Card) {
	// noop
}

func HumanAI() humanAI {
	return humanAI{}
}

type humanAI struct{}

func (ai humanAI) Bet(shuffled bool) int {
    if shuffled {
        fmt.Println("The deck was shuffled just now")
    }
	bet := 0
	fmt.Println("How much would you like to bet?")
	fmt.Scanf("%d\n", &bet)
	return bet
}

func (ai humanAI) Play(hand []deck.Card, dealer deck.Card) Move {
	var input string

	for {
		fmt.Println("Player: ", hand)
		fmt.Println("Dealer: ", dealer)
		fmt.Println("What would you like to do? (h)it, (s)tand, (d)ouble, s(p)lit")
		fmt.Scanf("%s\n", &input)

		switch input {
		case "h":
			return MoveHit
		case "s":
			return MoveStand
        case "d":
            return MoveDouble
        case "p":
            return MoveSplit
		default:
			fmt.Println("Invalid Option: ", input)
		}
	}
}

func (ai humanAI) Results(hands [][]deck.Card, dealer []deck.Card) {
	fmt.Println("\n==== FINAL HAND ====")
	fmt.Println("Player: ")
    for _, h := range hands {
        fmt.Println(" ", h)
    }
	fmt.Println("Dealer: ", dealer)
    fmt.Println()
}
