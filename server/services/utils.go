package services

import (
	"context"
	"fmt"
	"time"
)

type APITask struct {
	Label string
	Run   func() error
	Done  chan error 
}

// StartAPIScheduler est l'unique point d'accès à l'API externe.
// Une seule tâche API est exécutée à la fois, avec une pause fixe entre deux appels.
func StartAPIScheduler(ctx context.Context, tasks <-chan APITask) {
	const minSpacing = 7 * time.Second // marge de sécurité au-dessus de 10 req/min

	fmt.Printf("Scheduler API démarré | cadence = 1 requête / %s\n", minSpacing)

	for {
		select {
			case <-ctx.Done():
				fmt.Println("Arrêt du scheduler API")
				return

			case task := <-tasks:
				fmt.Printf("[API] Début tâche : %s\n", task.Label)

				err := task.Run()

				if task.Done != nil {
					task.Done <- err
					close(task.Done)
				}

				if err != nil {
					fmt.Printf("[API] Échec tâche : %s | erreur = %v\n", task.Label, err)
				} else {
					fmt.Printf("[API] Succès tâche : %s\n", task.Label)
				}

				select {
					case <-ctx.Done():
						fmt.Println("Arrêt du scheduler API")
						return
					case <-time.After(minSpacing):
				}
		}
	}
}