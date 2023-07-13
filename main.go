package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fatih/color"
)

const NumberOfPizzas = 10

var pizzasMade, pizzasFailed, total int

type Producer struct {
	data chan PizzaOrder
	quit chan chan error
}

type PizzaOrder struct {
	pizzaNumber int
	message     string
	success     bool
}

func (p *Producer) Close() error {
	ch := make(chan error)
	p.quit <- ch
	return <-ch
}

func makePizza(pizzaNumber int) *PizzaOrder {
	pizzaNumber++
	if pizzaNumber <= NumberOfPizzas {
		delay := rand.Intn(5) + 1
		fmt.Printf("received order #%d!\n", pizzaNumber)

		rnd := rand.Intn(12) + 1
		msg := ""
		success := false

		if rnd < 5 {
			pizzasFailed++
		} else {
			pizzasMade++
		}
		total++

		fmt.Printf("Making pizza #%d. It will take %d seconds....\n", pizzaNumber, delay)

		time.Sleep(time.Duration(delay) * time.Second)

		if rnd <= 2 {
			msg = fmt.Sprintf("*** we ran out ingredients for pizza #%d", pizzaNumber)
		} else if rnd <= 4 {
			msg = fmt.Sprintf("*** the cook is unavailable for pizza #%d", pizzaNumber)
		} else {
			success = true
			msg = fmt.Sprintf("Pizza #%d is ready", pizzaNumber)
		}
		p := PizzaOrder{
			pizzaNumber: pizzaNumber,
			message:     msg,
			success:     success,
		}
		return &p
	}

	return &PizzaOrder{
		pizzaNumber: pizzaNumber,
	}
}

func pizzeria(pizzaMaker *Producer) {
	// keep track of which pizza we are making
	var i = 0

	// run forever or until we receive a quit notification
	// try to make pizzas
	for {
		currentPizza := makePizza(i)
		if currentPizza != nil {
			i = currentPizza.pizzaNumber

			select {
			//we try to make a pizza (we sent smthn to the data channel)
			case pizzaMaker.data <- *currentPizza:
			case quitChan := <-pizzaMaker.quit:
				//close channel
				close(pizzaMaker.data)
				close(quitChan)
				return // exit the routine
			}
		}
	}
}

func main() {
	// seed the random number generator
	rand.NewSource(time.Now().UnixNano())

	// print out a message
	color.Cyan("The pizzeria is open for business")
	color.Cyan("_________________________________")

	// create a producer
	pizzaJob := &Producer{
		data: make(chan PizzaOrder),
		quit: make(chan chan error),
	}

	// run the producer background
	go pizzeria(pizzaJob)

	// create and run consumer
	for i := range pizzaJob.data {
		if i.pizzaNumber <= NumberOfPizzas {
			if i.success {
				color.Green(i.message)
				color.Green("Order #%d is out for delivery", i.pizzaNumber)
			} else {
				color.Red(i.message)
				color.Red("the customer is really mat")
			}
		} else {
			color.Cyan("DONE creating pizzaz")
			err := pizzaJob.Close()

			if err != nil {
				color.Red("*** error closing channel", err)
			}
		}
	}

	//print ending message
	color.Cyan("DONE FOR THE DAY")
	color.Cyan("We made %d pizzas, but faild to make %d with %d attempts in total", pizzasMade, pizzasFailed, total)

	switch {
	case pizzasFailed > 9:
		color.Red("it was an aweful day....")
	case pizzasFailed >= 6:
		color.Red("not a bad day")
	case pizzasFailed >= 4:
		color.Yellow("it was an ok day")
	case pizzasFailed >= 2:
		color.Yellow("it was an good day")
	default:
		color.Yellow("it was great  day")
	}
}
