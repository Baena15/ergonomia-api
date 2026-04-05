// Script para hacer administrador a un usuario
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL no está configurada")
	}

	if len(os.Args) < 2 {
		log.Fatal("Uso: go run make-admin.go <email>")
	}
	email := os.Args[1]

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}
	defer pool.Close()

	// Actualizar usuario a admin
	result, err := pool.Exec(context.Background(), 
		"UPDATE users SET is_admin = true WHERE email = $1", 
		email)
	if err != nil {
		log.Fatalf("Error actualizando usuario: %v", err)
	}

	if result.RowsAffected() == 0 {
		log.Fatalf("Usuario no encontrado: %s", email)
	}

	fmt.Printf("✅ Usuario %s ahora es administrador\n", email)
}
