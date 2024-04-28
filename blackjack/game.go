package blackjack

import (
	"errors"
	"fmt"

	deck "github.com/abdulkaderm36/gophercises-deck"
)

type state int8

const (
	stateBet state = iota
	statePlayerTurn
	stateDealerTurn
	stateHandOver
)

type hand struct {
	cards []deck.Card
	bet   int
}

type Game struct {
	nDecks          int
	nHands          int
	blackjackPayout float64

	state state
	deck  []deck.Card

	handIdx   int
	player    []hand
	playerBet int
	balance   int

	dealer   []deck.Card
	dealerAI AI
}

type Options struct {
	Decks           int
	Hands           int
	BlackjackPayout float64
}

func New(opts Options) Game {
	g := Game{
		state:    statePlayerTurn,
		dealerAI: dealerAI{},
		balance:  0,
	}

	if opts.Decks == 0 {
		opts.Decks = 3
	}
	if opts.Hands == 0 {
		opts.Hands = 10
	}
	if opts.BlackjackPayout == 0.0 {
		opts.BlackjackPayout = 1.5
	}
	g.nDecks = opts.Decks
	g.nHands = opts.Hands
	g.blackjackPayout = opts.BlackjackPayout

	return g
}

func (g *Game) currentHand() *[]deck.Card {
	switch g.state {
	case statePlayerTurn:
		return &g.player[g.handIdx].cards
	case stateDealerTurn:
		return &g.dealer
	default:
		panic("it isn't currently a player's turn")
	}
}

type Move func(g *Game) error

var (
	errBust   = errors.New("hand score exceeded 21")
	errDouble = errors.New("can only double on a hand with 2 cards")
)

func MoveHit(g *Game) error {
	hand := g.currentHand()

	var card deck.Card
	card, g.deck = draw(g.deck)
	*hand = append(*hand, card)

	if Score(*hand...) > 21 {
		return errBust
	}

	return nil
}

func MoveStand(g *Game) error {
	if g.state == stateDealerTurn {
		g.state++
		return nil
	}
	if g.state == statePlayerTurn {
		g.handIdx++
		if g.handIdx >= len(g.player) {
			g.state++
		}
		return nil
	}
	return errors.New("invalid state")
}

func MoveSplit(g *Game) error{
    cards := g.currentHand()

    if len(*cards) != 2 {
        return errors.New("you can only split with two cards in your hand")
    }
    if (*cards)[0].Rank != (*cards)[1].Rank {
        return errors.New("both cards must have the same rank to split")
    }
    
    g.player = append(g.player, hand{
            cards: []deck.Card{(*cards)[1]},
            bet:g.player[g.handIdx].bet, 
    })
    g.player[g.handIdx].cards = (*cards)[:1]
    return nil
}

func MoveDouble(g *Game) error {
	if len(*g.currentHand()) != 2 {
		return errDouble
	}
	g.playerBet *= 2
	MoveHit(g)
	return MoveStand(g)
}

func deal(g *Game) {
	playerHand := make([]deck.Card, 0, 5)
    g.handIdx = 0
	g.dealer = make([]deck.Card, 0, 5)

	var card deck.Card
	for i := 0; i < 2; i++ {
		card, g.deck = draw(g.deck)
		playerHand = append(playerHand, card)
		card, g.deck = draw(g.deck)
		g.dealer = append(g.dealer, card)
	}
	g.player = []hand{
		{
			cards: playerHand,
			bet:   g.playerBet,
		},
	}
	g.state = statePlayerTurn
}

func endRound(g *Game, ai AI) {
	dScore, dBlackjack := Score(g.dealer...), Blackjack(g.dealer...)
	allHands := make([][]deck.Card, len(g.player))
	for hi, hand := range g.player {
		cards := hand.cards
		allHands[hi] = cards
		pScore, pBlackjack := Score(cards...), Blackjack(cards...)
		winnings := hand.bet
		switch {
		case pBlackjack && dBlackjack:
			winnings = 0
		case dBlackjack:
			winnings = -winnings
		case pBlackjack:
			winnings = int(float64(winnings) * g.blackjackPayout)
		case pScore > 21:
			winnings = -winnings
		case dScore > 21:
			// win
		case pScore > dScore:
			// win
		case pScore < dScore:
			winnings = -winnings
		case pScore == dScore:
			winnings = 0
		}
		g.balance += winnings
	}

	ai.Results(allHands, g.dealer)
	g.player = nil
	g.dealer = nil
}

func bet(g *Game, ai AI, shuffled bool) {
	bet := ai.Bet(shuffled)
    if bet < 100 {
        panic("bet should be atleast 100")
    }
	g.playerBet = bet
}

func (g *Game) Play(ai AI) int {
	g.deck = nil
	min := 52 * g.nDecks / 3

	for i := 0; i < g.nHands; i++ {
		shuffled := false
		if len(g.deck) < min {
			g.deck = deck.New(deck.Deck(g.nDecks), deck.Shuffle)
			shuffled = true
		}
		bet(g, ai, shuffled)
		deal(g)

		if Blackjack(g.dealer...) {
			endRound(g, ai)
			continue
		}

		for g.state == statePlayerTurn {
			hand := make([]deck.Card, len(*g.currentHand()))
			copy(hand, *g.currentHand())
			move := ai.Play(hand, g.dealer[0])
			err := move(g)
			if err != nil {
				switch err {
				case errBust:
					MoveStand(g)
				case errDouble:
					fmt.Println(err.Error())
				default:
					panic(err)
				}
			}
		}

		for g.state == stateDealerTurn {
			hand := make([]deck.Card, len(g.dealer))
			copy(hand, g.dealer)
			move := g.dealerAI.Play(hand, g.dealer[0])
			move(g)
		}

		endRound(g, ai)
	}
	return g.balance
}

func draw(cards []deck.Card) (deck.Card, []deck.Card) {
	return cards[0], cards[1:]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minScore(hand ...deck.Card) int {
	var score int
	for _, card := range hand {
		score += min(int(card.Rank), 10)
	}

	return score
}

// Score will take in a hand of cards and return the best blackjack
// possible with the hand.
func Score(hand ...deck.Card) int {
	minScore := minScore(hand...)
	if minScore > 11 {
		return minScore
	}

	for _, card := range hand {
		if card.Rank == deck.Ace {
			return minScore + 10
		}
	}

	return minScore
}

func Blackjack(hand ...deck.Card) bool {
	return len(hand) == 2 && Score(hand...) == 21
}

// Soft will return true if the score of a hand is a soft score - that is if an Ace
// is being counted as 11 points
func Soft(hand ...deck.Card) bool {
	minScore := minScore(hand...)
	score := Score(hand...)
	return score != minScore
}
