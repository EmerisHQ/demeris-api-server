package database

import "fmt"

type ErrNoDestChain struct {
	Chain_a string
	Channel string
}

func (e ErrNoDestChain) Error() string {
	return fmt.Sprintf("no destination chain found for %s -> %s -> destination", e.Chain_a, e.Channel)
}

type ErrNoMatchingChannels struct{}

func (e ErrNoMatchingChannels) Error() string {
	return fmt.Sprintf("no matching channels found")
}
