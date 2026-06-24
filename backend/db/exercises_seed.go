package db

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

// SeedExercises checks if exercises table is empty and seeds 1,000 unique exercises
func SeedExercises(db *sql.DB) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&count)
	if err != nil {
		log.Fatalf("Error checking exercise count: %v", err)
	}

	if count >= 1000 {
		log.Printf("Exercises already seeded (%d exercises in DB)", count)
		return
	}

	log.Println("Seeding 1,000 unique exercises...")
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}

	stmt, err := tx.Prepare("INSERT INTO exercises (id, title, content, category, difficulty, is_endurance) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatalf("Error preparing statement: %v", err)
	}
	defer stmt.Close()

	// Seed counts
	totalCode := 250
	totalSpanish := 250
	totalEnglish := 250
	totalNumbers := 125
	totalSymbols := 125

	// Generate Code (Go, Java, Python, SQL, Angular)
	generateCodeExercises(stmt, totalCode, rng)

	// Generate Spanish texts
	generateSpanishExercises(stmt, totalSpanish, rng)

	// Generate English texts
	generateEnglishExercises(stmt, totalEnglish, rng)

	// Generate Numbers
	generateNumbersExercises(stmt, totalNumbers, rng)

	// Generate Symbols
	generateSymbolsExercises(stmt, totalSymbols, rng)

	if err := tx.Commit(); err != nil {
		log.Fatalf("Error committing transaction: %v", err)
	}

	var newCount int
	db.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&newCount)
	log.Printf("Successfully seeded %d exercises in database", newCount)
}

// Variables for generation
var (
	varNames     = []string{"user", "profile", "session", "metric", "achievement", "racha", "tropa", "manager", "data", "result", "config", "client", "server", "record"}
	funcNames    = []string{"getUser", "saveSession", "calculateStreak", "checkAchievements", "updateMetrics", "processData", "validateInput", "syncConfig", "fetchRecords"}
	structNames  = []string{"UserProfile", "TypingSession", "KeyMetric", "UserAchievement", "StreakTracker", "DatabaseConfig", "TropaMember"}
	types        = []string{"int", "string", "float64", "bool", "time.Time", "[]byte"}
	languages    = []string{"Golang", "Angular", "Java", "Python", "SQL"}
	topics       = []string{"mecanografía", "programación", "gimnasio y fuerza", "guitarra eléctrica", "música metal", "familia numerosa", "bases de datos", "despliegue en Docker"}
	musicians    = []string{"Dave Mustaine", "James Hetfield", "Björn Gelotte", "Johan Hegg", "LTD SC-207", "Line 6 Spider"}
	fitnessTerms = []string{"hipertrofia", "barra de 30kg", "mancuernas ajustables", "fuerza máxima", "volumen de entrenamiento"}
)

func generateCodeExercises(stmt *sql.Stmt, count int, rng *rand.Rand) {
	for i := 0; i < count; i++ {
		lang := languages[rng.Intn(len(languages))]
		var content, title, difficulty string
		isEndurance := 0

		// Alternate difficulties
		diffRoll := rng.Float64()
		if diffRoll < 0.4 {
			difficulty = "easy"
		} else if diffRoll < 0.8 {
			difficulty = "medium"
		} else {
			difficulty = "hard"
			if rng.Float64() < 0.3 {
				isEndurance = 1
			}
		}

		v1 := varNames[rng.Intn(len(varNames))]
		v2 := varNames[rng.Intn(len(varNames))]
		for v1 == v2 {
			v2 = varNames[rng.Intn(len(varNames))]
		}
		fn := funcNames[rng.Intn(len(funcNames))]
		st := structNames[rng.Intn(len(structNames))]
		tp := types[rng.Intn(len(types))]

		switch lang {
		case "Golang":
			title = fmt.Sprintf("Go: %s %d", st, i)
			if isEndurance == 1 {
				content = fmt.Sprintf(`package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type %s struct {
	ID        int       `+"`json:\"id\"`"+`
	Name      string    `+"`json:\"name\"`"+`
	Value     %s    `+"`json:\"value\"`"+`
	Active    bool      `+"`json:\"active\"`"+`
	CreatedAt time.Time `+"`json:\"created_at\"`"+`
}

func %s(ctx context.Context, db *sql.DB, %s int) (*%s, error) {
	query := "SELECT id, name, value, active, created_at FROM %ss WHERE id = ? AND active = 1"
	row := db.QueryRowContext(ctx, query, %s)

	var item %s
	err := row.Scan(&item.ID, &item.Name, &item.Value, &item.Active, &item.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active %s found for id %%d", %s)
		}
		return nil, fmt.Errorf("database query failure: %%w", err)
	}

	log.Printf("Successfully retrieved item %%d (%%s)", item.ID, item.Name)
	return &item, nil
}

func main() {
	fmt.Println("La Tropa backend running on Go 1.26")
}`, st, tp, fn, v1, st, strings.ToLower(st), v1, st, strings.ToLower(st), v1)
			} else {
				content = fmt.Sprintf(`func (s *Service) %s(ctx context.Context, %s %s) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.data[%s] = %s
		return nil
	}
}`, fn, v1, tp, v1, v2)
			}

		case "Angular":
			title = fmt.Sprintf("Angular TS: %sComponent %d", st, i)
			if isEndurance == 1 {
				content = fmt.Sprintf(`import { Component, signal, computed, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiService } from '../../services/api.service';

interface %s {
	id: number;
	name: string;
	xp: number;
}

@Component({
	selector: 'app-%s',
	standalone: true,
	imports: [CommonModule],
	templateUrl: './%s.component.html',
	styleUrls: ['./%s.component.css']
})
export class %sComponent implements OnInit {
	private api = inject(ApiService);
	
	readonly %s = signal<%s | null>(null);
	readonly hasXp = computed(() => {
		const current = this.%s();
		return current ? current.xp > 0 : false;
	});

	ngOnInit(): void {
		this.api.fetchProfile(%d).subscribe({
			next: (data) => this.%s.set(data),
			error: (err) => console.error('Error loading %s:', err)
		});
	}
}`, st, strings.ToLower(st), strings.ToLower(st), strings.ToLower(st), st, v1, st, v1, i%5+1, v1, v1)
			} else {
				content = fmt.Sprintf(`export class %sComponent {
	readonly %s = signal<%s>(%s);
	readonly doubled = computed(() => this.%s() + "_updated");

	updateValue(newValue: %s): void {
		this.%s.set(newValue);
	}
}`, st, v1, tp, getDefaultTSValue(tp), v1, tp, v1)
			}

		case "Java":
			title = fmt.Sprintf("Java: %sService %d", st, i)
			if isEndurance == 1 {
				content = fmt.Sprintf(`package com.latropa.typing.service;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

@Service
@Transactional
public class %sService {

	private final %sRepository repository;

	@Autowired
	public %sService(%sRepository repository) {
		this.repository = repository;
	}

	public List<%sDto> getActive%s() {
		return repository.findAllByActiveTrue().stream()
			.map(item -> new %sDto(item.getId(), item.getName()))
			.collect(Collectors.toList());
	}

	public Optional<%s> update%s(Long id, String name) {
		return repository.findById(id)
			.map(existing -> {
				existing.setName(name);
				return repository.save(existing);
			});
	}
}`, st, st, st, st, st, st, st, st, st)
			} else {
				content = fmt.Sprintf(`public List<String> %s(List<%s> items) {
	return items.stream()
		.filter(i -> i.isActive())
		.map(i -> i.getName().toUpperCase())
		.collect(Collectors.toList());
}`, fn, st)
			}

		case "Python":
			title = fmt.Sprintf("Python: %s %d", fn, i)
			if isEndurance == 1 {
				content = fmt.Sprintf(`import asyncio
import logging
from typing import List, Optional
from datetime import datetime

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class %sTracker:
    def __init__(self, name: str, threshold: int = 100):
        self.name = name
        self.threshold = threshold
        self.history: List[dict] = []

    async def add_record(self, value: float) -> bool:
        record = {
            "value": value,
            "timestamp": datetime.now().isoformat(),
            "alert": value > self.threshold
        }
        self.history.append(record)
        await asyncio.sleep(0.05)  # Simulate Async IO delay
        logger.info(f"Record added: {record}")
        return record["alert"]

    def get_alerts(self) -> List[dict]:
        return [r for r in self.history if r["alert"]]

async def main():
    tracker = %sTracker("TropaPerformance", threshold=75)
    for v in [55.2, 82.1, 91.5, 42.0]:
        is_alert = await tracker.add_record(v)
        if is_alert:
            print(f"Warning: {v} exceeded threshold!")`, st, st)
			} else {
				content = fmt.Sprintf(`def %s(%s: list, %s: int = 10) -> list:
    """Filter and limit elements based on criteria."""
    result = [x for x in %s if x > %s]
    return sorted(result)[:%s]`, fn, v1, v2, v1, v2, v2)
			}

		case "SQL":
			title = fmt.Sprintf("SQL Query %d", i)
			if isEndurance == 1 {
				content = fmt.Sprintf(`-- Complex Reporting Query for %s
WITH session_aggregates AS (
    SELECT 
        profile_id,
        COUNT(id) AS total_sessions,
        AVG(wpm) AS avg_wpm,
        AVG(accuracy) AS avg_accuracy,
        SUM(duration_seconds) AS total_seconds
    FROM sessions
    WHERE completed_at >= DATE('now', '-30 days')
    GROUP BY profile_id
),
key_error_rates AS (
    SELECT 
        profile_id,
        char,
        CAST(errors AS REAL) / CAST(attempts AS REAL) AS error_rate,
        ROW_NUMBER() OVER (PARTITION BY profile_id ORDER BY CAST(errors AS REAL) / CAST(attempts AS REAL) DESC) AS rank
    FROM key_metrics
    WHERE attempts > 20
)
SELECT 
    p.id,
    p.name,
    p.level,
    p.streak,
    sa.total_sessions,
    ROUND(sa.avg_wpm, 2) AS wpm_30_days,
    ROUND(sa.avg_accuracy * 100, 2) AS accuracy_percent,
    ker.char AS weakest_character,
    ROUND(ker.error_rate * 100, 2) AS character_error_rate
FROM profiles p
LEFT JOIN session_aggregates sa ON p.id = sa.profile_id
LEFT JOIN key_error_rates ker ON p.id = ker.profile_id AND ker.rank = 1
ORDER BY sa.avg_wpm DESC, p.name ASC;`, st)
			} else {
				content = fmt.Sprintf(`SELECT p.name, COUNT(s.id) AS total_runs, AVG(s.wpm) AS average_wpm
FROM profiles p
JOIN sessions s ON p.id = s.profile_id
WHERE s.completed_at >= DATE('now', '-%d days')
GROUP BY p.id
HAVING AVG(s.accuracy) > 0.95;`, i%30+1)
			}
		}

		id := fmt.Sprintf("code_%s_%d", strings.ToLower(lang), i)
		stmt.Exec(id, title, content, "code", difficulty, isEndurance)
	}
}

func generateSpanishExercises(stmt *sql.Stmt, count int, rng *rand.Rand) {
	templates := []string{
		"La práctica constante en la %s es vital para desarrollar memoria muscular. Tocar la %s requiere coordinación similar a programar en %s.",
		"En la familia %s, el orden y la paciencia son claves. Mantener una racha diaria de %s nos ayuda a todos a concentrarnos mejor y divertirnos.",
		"Entrenar con %s permite ganar fuerza e %s. Del mismo modo, escribir código limpio en %s requiere disciplina diaria.",
		"El guitarrista de metal practica sus solos a baja velocidad primero con su guitarra %s. En mecanografía, la precisión es más importante que la rapidez.",
		"Administrar sistemas mediante %s requiere conocer comandos precisos. Un solo error tipográfico puede tumbar la infraestructura del equipo.",
		"Los niños de la familia disfrutan aprendiendo cosas nuevas. Enseñar lógica mediante bloques de código en %s fomenta su creatividad.",
		"Desarrollar software en %s implica pensar en la estructura del código y los flujos de concurrencia para evitar problemas de sincronización.",
		"Una buena sesión de %s en casa, usando la %s y mancuernas ajustables, te deja listo para un día intenso de desarrollo de backend.",
	}

	for i := 0; i < count; i++ {
		isEndurance := 0
		var difficulty string
		diffRoll := rng.Float64()
		if diffRoll < 0.5 {
			difficulty = "easy"
		} else if diffRoll < 0.85 {
			difficulty = "medium"
		} else {
			difficulty = "hard"
			if rng.Float64() < 0.4 {
				isEndurance = 1
			}
		}

		var content strings.Builder
		topic := topics[rng.Intn(len(topics))]
		instrument := musicians[rng.Intn(len(musicians))]
		lang := languages[rng.Intn(len(languages))]
		fit := fitnessTerms[rng.Intn(len(fitnessTerms))]

		if isEndurance == 1 {
			// Long coherent text
			content.WriteString("Prueba de resistencia en Español. ")
			content.WriteString(fmt.Sprintf("Hoy hablaremos sobre %s y cómo se relaciona con el desarrollo de software moderno. ", topic))
			content.WriteString(fmt.Sprintf("Cuando tocamos un solo rápido con la guitarra %s, nuestro cerebro no piensa en notas individuales, sino en patrones motores. ", instrument))
			content.WriteString(fmt.Sprintf("Esto se conoce como memoria muscular y es idéntico a lo que ocurre cuando escribimos en %s o realizamos consultas SQL complejas en bases de datos relacionales. ", lang))
			content.WriteString(fmt.Sprintf("Para mantener un nivel óptimo de %s, se recomienda planificar bloques de entrenamiento físico de 3 a 4 días por semana. ", fit))
			content.WriteString("Esta disciplina física tiene un impacto directo en la concentración mental y la postura corporal al estar sentados frente al teclado. ")
			content.WriteString("En nuestra familia numerosa, apodada 'La Tropa', el ruido de los teclados mecánicos y las guitarras eléctricas es parte del día a día. ")
			content.WriteString("Cada uno de los niños tiene su propio ritmo de aprendizaje: algunos son más analíticos y lógicos, mientras que otros son sumamente enérgicos y creativos. ")
			content.WriteString("Esta diversidad neurodivergente en casa requiere que busquemos herramientas adaptables que ayuden a enfocar la atención y premien el esfuerzo diario mediante la gamificación. ")
			content.WriteString("Al final, ya sea levantando una barra de peso libre, ejecutando escalas en el traste de una Samick, o depurando un microservicio en contenedores Docker, la clave es la repetición deliberada. ")
			content.WriteString("Cometer errores es parte del proceso de aprendizaje, siempre y cuando analicemos dónde fallamos y volvamos a intentar el ejercicio con paciencia.")
		} else {
			tpl := templates[rng.Intn(len(templates))]
			if strings.Count(tpl, "%s") == 3 {
				content.WriteString(fmt.Sprintf(tpl, topic, instrument, lang))
			} else {
				content.WriteString(fmt.Sprintf(tpl, fit, instrument, lang))
			}
		}

		id := fmt.Sprintf("spa_%d", i)
		title := fmt.Sprintf("Español: Lección %d", i)
		stmt.Exec(id, title, content.String(), "spanish", difficulty, isEndurance)
	}
}

func generateEnglishExercises(stmt *sql.Stmt, count int, rng *rand.Rand) {
	templates := []string{
		"Regular practice in %s helps to build accurate finger patterns. Playing the riff from Megadeth requires high speed and precision.",
		"In the big family house, managing tasks is like organizing code in %s. Consistency is what keeps the daily streak going.",
		"Working out with %s develops strength. Eliciting clean code in %s demands a structured focus and attention to syntax errors.",
		"The electric guitar player dials in the tone on the %s. In touch typing, placing your hands on the home row is the initial step.",
		"Deploying containers via %s requires configuration files. Typing error-free parameters protects production servers from failing.",
		"Children learn logic and typing through fun gamified apps. Building interfaces in %s teaches them structure and layout principles.",
		"Writing structured queries in %s requires understanding database indexes, foreign keys, and relations to retrieve data quickly.",
		"A good session of %s using adjustable dumbbells keeps you energized for writing backend servers in Golang.",
	}

	for i := 0; i < count; i++ {
		isEndurance := 0
		var difficulty string
		diffRoll := rng.Float64()
		if diffRoll < 0.5 {
			difficulty = "easy"
		} else if diffRoll < 0.85 {
			difficulty = "medium"
		} else {
			difficulty = "hard"
			if rng.Float64() < 0.4 {
				isEndurance = 1
			}
		}

		var content strings.Builder
		topic := topics[rng.Intn(len(topics))]
		instrument := musicians[rng.Intn(len(musicians))]
		lang := languages[rng.Intn(len(languages))]
		fit := fitnessTerms[rng.Intn(len(fitnessTerms))]

		if isEndurance == 1 {
			// Long coherent text in English
			content.WriteString("English Endurance typing test. ")
			content.WriteString(fmt.Sprintf("Let us discuss the synergy between %s and software engineering principles. ", topic))
			content.WriteString(fmt.Sprintf("When performing a complex musical piece on a %s, the guitarist relies entirely on motor memory. ", instrument))
			content.WriteString(fmt.Sprintf("This is the exact same process used when writing efficient code in %s, or constructing clean database models. ", lang))
			content.WriteString(fmt.Sprintf("To maintain physical fitness and prevent repetitive strain injury (RSI), engaging in %s sessions is highly beneficial. ", fit))
			content.WriteString("It builds physical resilience and core strength, which directly improves posture during long hours of work. ")
			content.WriteString("In a large family environment like ours, balancing professional tasks, children, and personal hobbies requires planning. ")
			content.WriteString("Some kids are logical and highly focused on structural details, while others need interactive stimulus to maintain engagement. ")
			content.WriteString("Having a local gamified typing application allows everyone to progress at their own speed while earning rewards. ")
			content.WriteString("Whether you are lifting free weights, picking a heavy metal rhythm, or deploying microservices in Docker, repetition is key. ")
			content.WriteString("By tracking which specific characters cause the most latency or errors, we can dynamically build custom training sessions. ")
			content.WriteString("This deliberate practice turns weak spots into automated habits, paving the way for professional excellence.")
		} else {
			tpl := templates[rng.Intn(len(templates))]
			if strings.Count(tpl, "%s") == 3 {
				content.WriteString(fmt.Sprintf(tpl, topic, instrument, lang))
			} else {
				content.WriteString(fmt.Sprintf(tpl, fit, instrument, lang))
			}
		}

		id := fmt.Sprintf("eng_%d", i)
		title := fmt.Sprintf("English: Lesson %d", i)
		stmt.Exec(id, title, content.String(), "english", difficulty, isEndurance)
	}
}

func generateNumbersExercises(stmt *sql.Stmt, count int, rng *rand.Rand) {
	for i := 0; i < count; i++ {
		var difficulty string
		if i%3 == 0 {
			difficulty = "easy"
		} else if i%3 == 1 {
			difficulty = "medium"
		} else {
			difficulty = "hard"
		}

		var content string
		switch i % 5 {
		case 0:
			// IP addresses & ports
			content = fmt.Sprintf("192.168.1.%d, 10.0.0.%d:8080, 127.0.0.1:8081, 8.8.8.8, 1.1.1.1, %d.%d.%d.%d",
				rng.Intn(254)+1, rng.Intn(254)+1, rng.Intn(223)+1, rng.Intn(255), rng.Intn(255), rng.Intn(254)+1)
		case 1:
			// Dates and phone numbers
			content = fmt.Sprintf("2026-06-24, +56 9 9%d%d %d%d%d%d, 1984-11-23, 2018-12-07, 41 years old, family of 6",
				rng.Intn(9), rng.Intn(9), rng.Intn(9), rng.Intn(9), rng.Intn(9), rng.Intn(9))
		case 2:
			// Matrix of numbers (like database keys or rows)
			content = fmt.Sprintf("ID: %d00%d, Code: %d, XP: %d, Levels: 1 to 5, Streaks: %d days, weight: 30kg, 150W",
				rng.Intn(9)+1, rng.Intn(9), rng.Intn(900)+100, rng.Intn(5000)+1000, rng.Intn(30)+5)
		case 3:
			// Math formulas
			content = fmt.Sprintf("f(x) = %d*x^2 + %d*x - %d; y = %d / (%d + x); 3.14159 * r^2; 100%% accuracy",
				rng.Intn(9)+1, rng.Intn(9)+1, rng.Intn(20), rng.Intn(50), rng.Intn(5)+1)
		case 4:
			// Random phone/ID hashes
			content = fmt.Sprintf("9876543210 0123456789 555-019%d 8080 8081 30 7 207 150 13 14 7 7", rng.Intn(9))
		}

		id := fmt.Sprintf("num_%d", i)
		title := fmt.Sprintf("Números: Lección %d", i)
		stmt.Exec(id, title, content, "numbers", difficulty, 0)
	}
}

func generateSymbolsExercises(stmt *sql.Stmt, count int, rng *rand.Rand) {
	for i := 0; i < count; i++ {
		var difficulty string
		if i%3 == 0 {
			difficulty = "easy"
		} else if i%3 == 1 {
			difficulty = "medium"
		} else {
			difficulty = "hard"
		}

		var content string
		switch i % 5 {
		case 0:
			// Brackets, quotes, operators
			content = `if (a && b) { return c ? "yes" : "no"; } else if (a || !d) { print("error!"); }`
		case 1:
			// JSON format
			content = `{"status": 200, "data": {"user_id": 104, "metrics": [0.4, 0.98], "ok": true}}`
		case 2:
			// Complex symbol strings
			content = `!@#$%^&*()_+ -={}|[]\:";'<>?,./ ~` + "`" + ` § ±`
		case 3:
			// HTML tags & variables
			content = `<div class="container" id="kbd-overlay"> <span *ngIf="key() === '\''"> {{ key }} </span> </div>`
		case 4:
			// Pointer and channel ops
			content = `ch <- &User{Name: "Niko", XP: &xpVal}; val := <-ch; ptr := *val;`
		}

		id := fmt.Sprintf("sym_%d", i)
		title := fmt.Sprintf("Símbolos: Lección %d", i)
		stmt.Exec(id, title, content, "symbols", difficulty, 0)
	}
}

func getDefaultTSValue(tp string) string {
	switch tp {
	case "int":
		return "0"
	case "string":
		return "''"
	case "float64":
		return "0.0"
	case "bool":
		return "false"
	default:
		return "null"
	}
}
