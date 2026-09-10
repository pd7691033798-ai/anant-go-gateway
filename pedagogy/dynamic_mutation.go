package pedagogy

import (
	"fmt"
	"math/rand"
	"time"
)

type DynamicProblem struct {
	ProblemID     string
	Template      string
	CorrectAnswer int
	Options       []int
}

func GenerateMutatedMathProblem() DynamicProblem {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	speed := r.Intn(51) + 30
	duration := r.Intn(4) + 2
	distance := speed * duration

	template := fmt.Sprintf("एक रेलगाड़ी %d किमी/घंटे की गति से चल रही है। %d घंटे में दूरी तय करेगी?", speed, duration)
	options := []int{distance, distance + 20, distance - 15, distance + 35}
	r.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })

	return DynamicProblem{
		ProblemID:     fmt.Sprintf("DYN_%d", time.Now().Unix()),
		Template:      template,
		CorrectAnswer: distance,
		Options:       options,
	}
}
