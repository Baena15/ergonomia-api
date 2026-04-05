// Script para ejecutar migraciones SQL
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Obtener DATABASE_URL de variable de entorno
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL no está configurada")
	}

	// Conectar a la base de datos
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}
	defer pool.Close()

	// Verificar conexión
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Error haciendo ping a la base de datos: %v", err)
	}
	fmt.Println("✅ Conectado a la base de datos")

	// Leer archivo de migración
	migrationPath := filepath.Join("..", "migrations", "001_init.sql")
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		// Intentar ruta alternativa
		migrationPath = filepath.Join("migrations", "001_init.sql")
		content, err = os.ReadFile(migrationPath)
		if err != nil {
			log.Fatalf("Error leyendo archivo de migración: %v", err)
		}
	}

	// Dividir en statements (separados por ;)
	statements := strings.Split(string(content), ";")

	// Ejecutar cada statement
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		_, err := pool.Exec(context.Background(), stmt)
		if err != nil {
			// Algunos statements pueden fallar si ya existen (como CREATE EXTENSION IF NOT EXISTS)
			// Continuamos con el siguiente
			fmt.Printf("⚠️  Statement %d: %v\n", i+1, err)
		} else {
			fmt.Printf("✅ Statement %d ejecutado\n", i+1)
		}
	}

	fmt.Println("\n🎉 Migraciones completadas!")
}
