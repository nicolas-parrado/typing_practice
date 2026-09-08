package db

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

type TempExercise struct {
	Title       string
	Content     string
	Category    string
	Difficulty  string
	IsEndurance int
	Stage       string
	TargetWPM   float64
	MinAccuracy float64
	SourceIDs   []int
}

// SeedExercises checks if exercises table is empty or stale, and seeds progressive lessons
func SeedExercises(db *sql.DB) {
	// 1. Detect if exercises table has old non-progressive seeds or legacy Spanish-layout home row
	var hasStages bool
	_ = db.QueryRow("SELECT EXISTS(SELECT 1 FROM exercises WHERE stage IS NOT NULL AND stage != '')").Scan(&hasStages)
	var hasLegacyHomeRow bool
	_ = db.QueryRow("SELECT EXISTS(SELECT 1 FROM exercises WHERE id = 'spa_1' AND content LIKE '%jklñ%')").Scan(&hasLegacyHomeRow)

	if !hasStages || hasLegacyHomeRow {
		log.Println("Old exercises detected without progressive stages or using non-US-International home row. Purging exercises table for US International curriculum...")
		_, err := db.Exec("DELETE FROM exercises")
		if err != nil {
			log.Printf("Error purging legacy exercises: %v", err)
		}
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&count)
	if err != nil {
		log.Fatalf("Error checking exercise count: %v", err)
	}

	if count >= 700 {
		log.Printf("Exercises already seeded (%d progressive exercises in DB)", count)
		return
	}

	log.Println("Generating progressive structured curriculum in memory...")
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	var tempExercises []TempExercise
	tempExercises = append(tempExercises, generateSpanishExercises(rng)...)
	tempExercises = append(tempExercises, generateEnglishExercises(rng)...)
	tempExercises = append(tempExercises, generateSymbolsExercises(rng)...)
	tempExercises = append(tempExercises, generateCodeExercises(rng)...)
	tempExercises = append(tempExercises, generateNumbersExercises(rng)...)

	// Group by category
	categorized := make(map[string][]TempExercise)
	for _, ex := range tempExercises {
		categorized[ex.Category] = append(categorized[ex.Category], ex)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO exercises 
		(id, title, content, category, difficulty, is_endurance, stage, target_wpm, min_accuracy) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		log.Fatalf("Error preparing statement: %v", err)
	}
	defer stmt.Close()

	for cat, list := range categorized {
		for index, ex := range list {
			seq := index + 1
			var id, finalTitle string

			switch cat {
			case "spanish":
				id = fmt.Sprintf("spa_%d", seq)
				finalTitle = fmt.Sprintf("Español: Lección %d", seq)
			case "english":
				id = fmt.Sprintf("eng_%d", seq)
				finalTitle = fmt.Sprintf("English: Lesson %d", seq)
			case "symbols":
				id = fmt.Sprintf("sym_%d", seq)
				finalTitle = fmt.Sprintf("Símbolos: Lección %d", seq)
			case "code":
				id = fmt.Sprintf("code_%d", seq)
				finalTitle = fmt.Sprintf("Código: Lección %d", seq)
			case "numbers":
				id = fmt.Sprintf("num_%d", seq)
				finalTitle = fmt.Sprintf("Números: Lección %d", seq)
			default:
				id = fmt.Sprintf("ex_%s_%d", cat, seq)
				finalTitle = fmt.Sprintf("Lección %d", seq)
			}

			_, err = stmt.Exec(id, finalTitle, ex.Content, ex.Category, ex.Difficulty, ex.IsEndurance, ex.Stage, ex.TargetWPM, ex.MinAccuracy)
			if err != nil {
				log.Fatalf("Error inserting exercise %s: %v", id, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Error committing transaction: %v", err)
	}

	var newCount int
	db.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&newCount)
	log.Printf("Successfully seeded %d progressive exercises with adaptive targets in database", newCount)
}

// -------------------------------------------------------------
// SPANISH PROGRESSION (250 lessons)
// -------------------------------------------------------------
func generateSpanishExercises(rng *rand.Rand) []TempExercise {
	var list []TempExercise

	// Etapa 1: Fila Base (Home Row EEUU Internacional: a s d f j k l ;) - 15 lecciones
	homeRowPatterns := []string{
		"asdf jkl; asdf jkl; asdf jkl; fdsa ;lkj fdsa ;lkj",
		"aaa sss ddd fff jjj kkk lll ;;; fff jjj ddd kkk",
		"fa da la sa ja ka ha ga ;a fa da la sa ja ka ha ga ;a",
		"fad kas sal fal das lak gas fas dak sal kas",
		"la sal da la gala salsa jala alas alas gala",
		"faja faja salsa salsa falda falda alada alada",
		"da la sal a la gala da la falda a la salsa",
		"gala jala la falda hada faja jala sala gas",
		"alfalfa fajas saladas aladas fallas salsa sal",
		"alas aladas dadas a la sala jala la sal sala",
		"la gala jala la salsa salada a la falda alada",
		"da la faja a la gala salsa salada da la sala",
		"falla la salsa jala la gala salada a la sala",
		"faldas aladas dadas a la salsa jala la gala",
		"la sala da sal a la gala alada con su faja",
	}
	for i, p := range homeRowPatterns {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Fila Base %d", i+1),
			Content:     p,
			Category:    "spanish",
			Difficulty:  "easy",
			IsEndurance: 0,
			Stage:       "Fila Base",
			TargetWPM:   18.0,
			MinAccuracy: 0.90,
		})
	}

	// Etapa 2: Filas Superior e Inferior (e, i, r, u, t, o, c, m, p, v, b, n) - 25 lecciones
	shortWordsSpanish := []string{
		"el tren corre por la via del norte con paso firme",
		"la mesa de pino tiene una taza de cafe caliente",
		"mi perro corre tras la pelota en el patio verde",
		"el sol brilla sobre el rio claro de la colina",
		"un buen libro abre la mente y calma la rutina",
		"la casa tiene una ventana que mira hacia el mar",
		"el viento mueve las hojas secas del viejo roble",
		"beber agua fresca cada manana da vitalidad y salud",
		"el camino de tierra sube lento por la montana",
		"un plato de sopa caliente reconforta en el invierno",
		"la musica suave llena la sala de paz y calma",
		"el barco navega seguro bajo la luz de la luna",
		"cada paso firme acerca la meta con paso seguro",
		"el reloj de pared marca las horas sin detenerse",
		"un salto largo sobre la arena moja los pies",
		"la luz del faro guia los barcos en la noche oscura",
		"el puente de piedra cruza el rio con paso firme",
		"un canto de ave saluda el nuevo amanecer en el campo",
		"el aroma del pan recien horneado llena la cocina",
		"la lluvia suave cae sobre el tejado de madera",
		"una taza de te caliente ayuda a descansar en paz",
		"el fuego de la chimenea calienta toda la cabana",
		"el gato duerme plácido sobre el sillon de cuero",
		"caminar bajo los arboles verdes renueva la energia",
		"el cielo azul se despeja tras la tormenta de verano",
	}
	for i, w := range shortWordsSpanish {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Filas y Alcances %d", i+1),
			Content:     w,
			Category:    "spanish",
			Difficulty:  "easy",
			IsEndurance: 0,
			Stage:       "Filas y Alcances",
			TargetWPM:   20.0,
			MinAccuracy: 0.90,
		})
	}

	// Etapa 3: Mayúsculas, Tildes y Puntuación - 35 lecciones
	accentedSpanish := []string{
		"Papá tomó café recién hecho en la mañana, mientras leía las noticias.",
		"Sofía corrió más rápido que el viento para alcanzar el autobús escolar.",
		"¿Dónde dejaste el teléfono móvil? Está sonando sobre la mesa del comedor.",
		"Nicolás practica con su guitarra eléctrica todos los días con gran pasión.",
		"El éxito requiere paciencia, dedicación constante y mucha disciplina personal.",
		"¡Qué alegría verte de nuevo por aquí! Espero que todo marche de maravilla.",
		"La educación es el arma más poderosa para transformar el futuro del mundo.",
		"Joyce organiza la agenda familiar con gran atención y cuidado en cada detalle.",
		"Pablo descubrió una solución ingeniosa para resolver el acertijo de matemáticas.",
		"Fernando prefiere entrenar temprano para aprovechar al máximo las horas del día.",
		"Luciano tiene una memoria visual increíble y un pensamiento lógico admirable.",
		"El océano Pacífico baña las costas con olas gigantescas y aguas profundas.",
		"Siempre es útil guardar una copia de seguridad antes de modificar un archivo crítico.",
		"¿Podrías decirme qué hora es, por favor? Se me detuvo el reloj de pulsera.",
		"La perseverancia vence lo que la dicha no alcanza en cualquier meta difícil.",
		"El volcán arrojó cenizas durante la madrugada, alertando a toda la población.",
		"Escribir con buena ortografía y precisión es una muestra de respeto al lector.",
		"Cada árbol en el bosque cumple una función vital para el equilibrio del planeta.",
		"¡Nunca te rindas cuando el camino se ponga empinado! Respira hondo y continúa.",
		"El desarrollo ágil promueve entregas rápidas y una comunicación constante y clara.",
		"La melodía de la guitarra resonó en el teatro con un sonido nítido y potente.",
		"Comer frutas frescas y verduras variadas aporta los nutrientes que el cuerpo necesita.",
		"¿Cuánto tiempo tardará en compilarse el proyecto completo en el servidor central?",
		"La tecnología avanza a un ritmo vertiginoso, abriendo posibilidades extraordinarias.",
		"Un buen descanso nocturno es fundamental para fijar los aprendizajes del día a día.",
		"El café de origen colombiano tiene un aroma suave y un sabor inconfundible.",
		"Los niños jugaban en el parque mientras el sol se ocultaba en el horizonte.",
		"La biblioteca municipal conserva documentos históricos de gran valor cultural.",
		"Mantener la calma en momentos de tensión ayuda a tomar decisiones acertadas.",
		"El tren de alta velocidad redujo el trayecto entre ambas ciudades a la mitad.",
		"La curiosidad científica impulsa el progreso humano hacia nuevos descubrimientos.",
		"Aprender a tocar un instrumento musical fomenta la neuroplasticidad cerebral.",
		"La honestidad y la empatía son valores indispensables en cualquier comunidad.",
		"El invierno trajo nevadas abundantes en las cumbres más altas de la cordillera.",
		"La práctica deliberada convierte las acciones complejas en reflejos automáticos.",
	}
	for i, a := range accentedSpanish {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Acentos y Puntuación %d", i+1),
			Content:     a,
			Category:    "spanish",
			Difficulty:  "medium",
			IsEndurance: 0,
			Stage:       "Acentos y Puntuación",
			TargetWPM:   22.0,
			MinAccuracy: 0.90,
		})
	}

	// Etapa 4: Palabras de Alta Frecuencia (Automatización Rítmica) - 45 lecciones
	highFreqSpanishPool := []string{
		"de la que el en y a los se del las por un para con no una su al lo como mas pero sus le ya o fue este ha si porque esta son entre cuando muy sin sobre ser tiene tambien me hasta hay donde todo era estos vida tiempo uno nuevo otro solo ahora cada parte lugar hombre pais mismo mundo trabajo mayor primer mano fin punto caso estado semana grupo hecho forma caso general propio luego agua orden lado noche camino",
	}
	words := strings.Fields(highFreqSpanishPool[0])
	for i := 0; i < 45; i++ {
		start := (i * 7) % (len(words) - 15)
		chunk := strings.Join(words[start:start+12], " ")
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Palabras Frecuentes %d", i+1),
			Content:     chunk,
			Category:    "spanish",
			Difficulty:  "medium",
			IsEndurance: 0,
			Stage:       "Palabras Frecuentes",
			TargetWPM:   26.0,
			MinAccuracy: 0.92,
		})
	}

	// Etapa 5: Fluidez y Oraciones Completas - 80 lecciones
	for i := 0; i < 80; i++ {
		idx1 := (i * 3) % len(spanishSentences)
		idx2 := (i*3 + 1) % len(spanishSentences)
		content := fmt.Sprintf("%s %s", spanishSentences[idx1], spanishSentences[idx2])
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Fluidez Sintáctica %d", i+1),
			Content:     content,
			Category:    "spanish",
			Difficulty:  "medium",
			IsEndurance: 0,
			Stage:       "Fluidez Sintáctica",
			TargetWPM:   30.0,
			MinAccuracy: 0.92,
			SourceIDs:   []int{idx1, idx2},
		})
	}

	// Etapa 6: Velocidad y Resistencia - 50 lecciones
	for i := 0; i < 50; i++ {
		idx1 := (i * 4) % len(spanishSentences)
		idx2 := (i*4 + 1) % len(spanishSentences)
		idx3 := (i*4 + 2) % len(spanishSentences)
		idx4 := (i*4 + 3) % len(spanishSentences)
		content := fmt.Sprintf("%s %s %s %s", spanishSentences[idx1], spanishSentences[idx2], spanishSentences[idx3], spanishSentences[idx4])
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Velocidad y Resistencia %d", i+1),
			Content:     content,
			Category:    "spanish",
			Difficulty:  "hard",
			IsEndurance: 1,
			Stage:       "Velocidad y Resistencia",
			TargetWPM:   38.0,
			MinAccuracy: 0.94,
			SourceIDs:   []int{idx1, idx2, idx3, idx4},
		})
	}

	return list
}

// -------------------------------------------------------------
// ENGLISH PROGRESSION (250 lessons)
// -------------------------------------------------------------
func generateEnglishExercises(rng *rand.Rand) []TempExercise {
	var list []TempExercise

	// Stage 1: Home Row (15 lessons)
	homeRowEn := []string{
		"asdf jkl; asdf jkl; asdf jkl; fdsa ;lkj fdsa ;lkj",
		"aaa sss ddd fff jjj kkk lll ;;; fff jjj ddd kkk",
		"fa da la sa ja ka ;a fa da la sa ja ka ;a",
		"fad kas sal fal das lak ;al fas dak sal kas",
		"all dads ask for salad all lads had a flask",
		"a sad lad falls as a glad dad asks for a flask",
		"ask dad for a salad and a flask of fall soda",
		"dads add salads as lads fall all glad lads ask",
		"a flask for a sad dad and a salad for all lads",
		"glad dads ask sad lads as all fall salads add",
		"all lads ask for a salad as dad adds a flask",
		"sad dads fall as glad lads add salads to flasks",
		"a salad falls as dad asks for all glad flasks",
		"add a salad ask dad for a flask and fall lad",
		"glad lads and sad dads ask for all fall salads",
	}
	for i, p := range homeRowEn {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Home Row %d", i+1),
			Content:     p,
			Category:    "english",
			Difficulty:  "easy",
			IsEndurance: 0,
			Stage:       "Home Row",
			TargetWPM:   18.0,
			MinAccuracy: 0.90,
		})
	}

	// Stage 2: Top & Bottom Extensions (25 lessons)
	extensionsEn := []string{
		"the quick brown fox jumps over the lazy dog in the park",
		"pack my box with five dozen liquor jugs before noon",
		"bright red apples grow on tall trees near the clean lake",
		"we walk along the quiet river under the morning sun",
		"a warm cup of tea brings comfort on a chilly autumn day",
		"she read a fascinating book about modern computing and space",
		"the cold rain fell steadily on the roof through the night",
		"birds sing cheerful melodies high up in the green branches",
		"he wrote clean code and tested each unit before deploying",
		"the train arrived right on schedule at the central platform",
		"fresh bread from the local bakery smells delicious every morning",
		"deep focus allows engineers to solve intricate system challenges",
		"they climbed the rocky hill to see the sunset over the valley",
		"learning touch typing requires steady repetition and patience",
		"the ocean breeze carries the soothing sound of distant waves",
		"good posture and a comfortable chair prevent wrist fatigue",
		"a gentle fire warms the room while cold snow falls outside",
		"he tuned his electric guitar carefully before starting the solo",
		"the garden blooms with colorful flowers under the blue sky",
		"consistent daily practice turns conscious effort into muscle memory",
		"they built an automated pipeline using docker and continuous testing",
		"the library offers a peaceful place to study and concentrate",
		"clear thinking leads to maintainable code and elegant software designs",
		"the stars shine brightly on a dark clear winter night",
		"every small improvement in typing speed compounds significantly over time",
	}
	for i, w := range extensionsEn {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Row Extensions %d", i+1),
			Content:     w,
			Category:    "english",
			Difficulty:  "easy",
			IsEndurance: 0,
			Stage:       "Row Extensions",
			TargetWPM:   20.0,
			MinAccuracy: 0.90,
		})
	}

	// Stage 3: Shift & Punctuation (35 lessons)
	punctEn := []string{
		"Do you know where the server logs are stored? Check the /var/log directory.",
		"Always write unit tests, handle errors gracefully, and document your public APIs.",
		"Can you believe how fast this Go web service compiles and executes under load?",
		"Keep your hands relaxed, float your wrists, and maintain a steady typing tempo.",
		"What a remarkable performance! The guitarist played a flawless seven-string solo.",
		"Remember: premature optimization is the root of all evil in software design.",
		"Have you configured Docker Compose properly? Don't forget the environment variables.",
		"Every morning, Niko drinks a cup of black coffee and reviews pull requests.",
		"Is PostgreSQL running on localhost:5432? Verify the connection string right away.",
		"Success is not final, failure is not fatal; it is the courage to continue that counts.",
		"Why choose microservices when a modular monolith can easily handle your current scale?",
		"Pay attention to accuracy first! Speed naturally follows once muscle memory is solid.",
		"The mechanical keyboard features tactile brown switches and seamless keycap profiles.",
		"Did you commit your changes to git? Use: git commit -m 'feat: add user profile'.",
		"Patience and emotional support are the most important anchors in a large family.",
		"Could you explain how goroutines coordinate communication using buffered channels?",
		"Heavy barbell workouts combined with sound nutrition stimulate optimal hypertrophy.",
		"Never stop learning: new technologies emerge rapidly, but solid fundamentals endure.",
		"Does Angular's reactive signal system simplify state management in large frontends?",
		"Great software is built by empathetic teams who prioritize user needs above all.",
		"Are you ready for the endurance test? Keep your rhythm steady and breathe normally.",
		"The terminal prompt is ready: type 'go test ./...' to run the automated test suite.",
		"When refactoring legacy code, always ensure comprehensive test coverage exists first.",
		"Heavy metal music provides high energy and intense focus during strenuous training sessions.",
		"Is there anything more satisfying than typing a complex program without a single typo?",
		"Simplicity is prerequisite for reliability, as Edsger Dijkstra famously remarked.",
		"How much memory does this microservice consume when processing ten thousand requests?",
		"Focus on pressing each key cleanly without tensing your shoulders or forearms.",
		"Continuous integration ensures that breaking changes are caught before reaching production.",
		"The sound of mechanical switches clicking rhythmically creates an enjoyable flow state.",
		"Can you find the bug in this function? It seems to be an off-by-one indexing error.",
		"Consistency in strength training, three or four sessions weekly, yields lasting gains.",
		"Good typography, clean contrast, and generous whitespace elevate modern web applications.",
		"What is the average latency of your database queries under peak concurrent traffic?",
		"Celebrate every small milestone as you progress toward your target typing speed.",
	}
	for i, p := range punctEn {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Shift & Punctuation %d", i+1),
			Content:     p,
			Category:    "english",
			Difficulty:  "medium",
			IsEndurance: 0,
			Stage:       "Shift & Punctuation",
			TargetWPM:   22.0,
			MinAccuracy: 0.90,
		})
	}

	// Stage 4: Common Words (45 lessons)
	commonEnWords := []string{
		"the of and to a in is you that it he was for on are as with his they I at be this have from or one had by word but not what all were we when your can said there use an each which she do how their if will up other about out many then them these so some her would make like him into time has look two more write go see number no way could people my than first water been call who oil its now find long down day did get come made may part",
	}
	enWords := strings.Fields(commonEnWords[0])
	for i := 0; i < 45; i++ {
		start := (i * 7) % (len(enWords) - 15)
		chunk := strings.Join(enWords[start:start+12], " ")
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Common Words %d", i+1),
			Content:     chunk,
			Category:    "english",
			Difficulty:  "medium",
			IsEndurance: 0,
			Stage:       "Common Words",
			TargetWPM:   26.0,
			MinAccuracy: 0.92,
		})
	}

	// Stage 5: Fluency & Rhythm (80 lessons)
	for i := 0; i < 80; i++ {
		idx1 := (i * 3) % len(englishSentences)
		idx2 := (i*3 + 1) % len(englishSentences)
		content := fmt.Sprintf("%s %s", englishSentences[idx1], englishSentences[idx2])
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Fluency & Rhythm %d", i+1),
			Content:     content,
			Category:    "english",
			Difficulty:  "medium",
			IsEndurance: 0,
			Stage:       "Fluency & Rhythm",
			TargetWPM:   30.0,
			MinAccuracy: 0.92,
			SourceIDs:   []int{idx1, idx2},
		})
	}

	// Stage 6: Speed & Endurance (50 lessons)
	for i := 0; i < 50; i++ {
		idx1 := (i * 4) % len(englishSentences)
		idx2 := (i*4 + 1) % len(englishSentences)
		idx3 := (i*4 + 2) % len(englishSentences)
		idx4 := (i*4 + 3) % len(englishSentences)
		content := fmt.Sprintf("%s %s %s %s", englishSentences[idx1], englishSentences[idx2], englishSentences[idx3], englishSentences[idx4])
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Speed & Endurance %d", i+1),
			Content:     content,
			Category:    "english",
			Difficulty:  "hard",
			IsEndurance: 1,
			Stage:       "Speed & Endurance",
			TargetWPM:   38.0,
			MinAccuracy: 0.94,
			SourceIDs:   []int{idx1, idx2, idx3, idx4},
		})
	}

	return list
}

// -------------------------------------------------------------
// SYMBOLS PROGRESSION (75 lessons) - No more regex walls!
// -------------------------------------------------------------
func generateSymbolsExercises(rng *rand.Rand) []TempExercise {
	var list []TempExercise

	// Etapa 1: Delimitadores Básicos - 15 lecciones (Target: 10 WPM, 88% Acc)
	delimiters := []string{
		"(a, b); (x, y); (item, index); (key, value); (req, res);",
		"[0, 1]; [10, 20]; [item]; [true, false]; [first, last];",
		"{ id: 1 }; { name: 'Niko' }; { ok: true }; { xp: 100 };",
		"(a + b); [x, y, z]; { count: 0 }; (foo, bar); [head, tail];",
		"func(a, b); list[0]; obj.prop; data[key]; run(true, false);",
		"fn(1, 2); arr[3]; { status: 200 }; log(msg); res[idx];",
		"items[0]; keys[1]; { active: true }; calc(a, b, c);",
		"call(data); [x, y]; { level: 1 }; list[i]; send(packet);",
		"(10, 20, 30); [a, b, c]; { flag: false }; emit(event);",
		"query(sql); [val]; { port: 8080 }; get(id); store[name];",
		"(req, res, next); [1, 2, 3]; { role: 'admin' }; save(record);",
		"fetch(url); [row, col]; { width: 100 }; parse(input);",
		"(x, y); [idx]; { ready: true }; push(val); splice(0, 1);",
		"init(config); [top, bottom]; { debug: false }; check(flag);",
		"(a, b); [start, end]; { count: 10 }; reset(); finish();",
	}
	for i, d := range delimiters {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Delimitadores Básicos %d", i+1),
			Content:     d,
			Category:    "symbols",
			Difficulty:  "easy",
			IsEndurance: 0,
			Stage:       "Delimitadores Básicos",
			TargetWPM:   10.0,
			MinAccuracy: 0.88,
		})
	}

	// Etapa 2: Operadores y Asignación - 15 lecciones (Target: 12 WPM, 88% Acc)
	operators := []string{
		"x = 10; y = 20; total = x + y; diff = y - x;",
		"a == b; c != d; x === y; count += 1; balance -= 10;",
		"ratio = total / count; product = a * b; rest = x % 2;",
		"val := 100; name := 'Go'; sum += val; diff -= 5;",
		"x = x + 1; y = y * 2; z = z / 4; rem = n % 10;",
		"i += 1; j -= 1; k *= 2; score += 15; tries += 1;",
		"a = b + c; d = e - f; g = h * i; j = k / l;",
		"x := a + b; y := c * d; count += 10; ok := true;",
		"ans = (a + b) * c; res = (x - y) / 2; val += 100;",
		"total := count * price; tax := total * 0.19; final = total + tax;",
		"p1 = (x1, y1); p2 = (x2, y2); dist = (x2 - x1) + (y2 - y1);",
		"a == 0; b != 1; c === null; x !== undefined; val := 42;",
		"ptr := &val; ref := *ptr; count += 1; flag = true;",
		"step += 2; speed -= 5; zoom *= 1.5; scale /= 2;",
		"cost = base + (hours * rate); net = cost - discount;",
	}
	for i, o := range operators {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Operadores y Asignación %d", i+1),
			Content:     o,
			Category:    "symbols",
			Difficulty:  "easy",
			IsEndurance: 0,
			Stage:       "Operadores y Asignación",
			TargetWPM:   12.0,
			MinAccuracy: 0.88,
		})
	}

	// Etapa 3: Lógica y Flechas - 15 lecciones (Target: 14 WPM, 90% Acc)
	logicArrows := []string{
		"if (a > 0 && b < 10) { return true; }",
		"if (user != null || isAdmin) { allow(); }",
		"const fn = (x) => x * 2; const add = (a, b) => a + b;",
		"const filter = (item) => item.active && item.score >= 50;",
		"if (!isValid && count > 0) { throw new Error('invalid'); }",
		"const isEven = (n) => (n % 2 === 0);",
		"while (i < len && !found) { if (arr[i] === target) found = true; }",
		"const status = isReady ? 'DONE' : 'PENDING';",
		"arr.map((x) => x + 1).filter((x) => x > 5);",
		"return a !== null && b !== undefined ? a : b;",
		"const render = (props) => <div>{props.label}</div>;",
		"if (level >= 10 && xp >= 5000) { unlockBadge('master'); }",
		"const clamp = (val, min, max) => Math.min(Math.max(val, min), max);",
		"ch <- &Data{ID: 1}; msg := <-ch;",
		"const result = valid ? (x > y ? x : y) : 0;",
	}
	for i, l := range logicArrows {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Operadores Lógicos y Flechas %d", i+1),
			Content:     l,
			Category:    "symbols",
			Difficulty:  "medium",
			IsEndurance: 0,
			Stage:       "Operadores Lógicos y Flechas",
			TargetWPM:   14.0,
			MinAccuracy: 0.90,
		})
	}

	// Etapa 4: Comillas, Template Literals y Selectores - 15 lecciones (Target: 16 WPM, 90% Acc)
	stringsSelectors := []string{
		`const msg = "Hola Mundo"; const user = 'Niko';`,
		`const url = ` + "`https://api.dev/users/${id}/profile`;",
		`const greeting = ` + "`Welcome, ${name}! Your level is ${level}.`;",
		`const query = "#main-container .btn-primary[disabled]";`,
		`const css = ` + "`color: ${theme.color}; font-size: 16px;`;",
		`curl -X POST -H "Content-Type: application/json" http://localhost:8080`,
		`git commit -m "feat: add adaptive progression targets"`,
		`docker run -d -p 8080:8080 --name backend-api my-app:latest`,
		`const [user, setUser] = useState<User | null>(null);`,
		`const { id, name, ...rest } = activeProfile();`,
		`const paths = glob.sync('**/src/**/*.spec.ts');`,
		`const config = { host: process.env.DB_HOST || 'localhost', port: 5432 };`,
		`const endpoint = ` + "`/api/v1/sessions?profile_id=${profileId}`;",
		`assert client.get("/health").status_code == 200`,
		`sudo systemctl restart docker.service && tail -f /var/log/syslog`,
	}
	for i, s := range stringsSelectors {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Comillas y Selectores %d", i+1),
			Content:     s,
			Category:    "symbols",
			Difficulty:  "medium",
			IsEndurance: 0,
			Stage:       "Comillas y Selectores",
			TargetWPM:   16.0,
			MinAccuracy: 0.90,
		})
	}

	// Etapa 5: Sintaxis Mixta de Programación - 10 lecciones (Target: 18 WPM, 90% Acc)
	mixedSyntax := []string{
		`array.filter(x => x !== null).map(x => x.trim().toUpperCase());`,
		`[ngClass]="{'active': isActive(), 'disabled': !isEnabled()}"`,
		`{"status": 200, "ok": true, "data": {"user_id": 101, "role": "admin"}}`,
		`const average = arr.reduce((acc, curr) => acc + curr, 0) / arr.length;`,
		`func (s *Service) Load(ctx context.Context) (<-chan *Data, error)`,
		`struct KeyMetric { attempts: i32, errors: i32, latency_ms: f32 }`,
		`const result = await Promise.all([task1(), task2(), task3()]);`,
		`type ProfileMap = Record<string, { xp: number; streak: number }>;`,
		`SELECT * FROM profiles WHERE streak >= 7 AND sound_enabled = 1;`,
		`<div class="card" [class.locked]="!unlocked"> <app-badge [id]="id" /> </div>`,
	}
	for i, m := range mixedSyntax {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Sintaxis Mixta %d", i+1),
			Content:     m,
			Category:    "symbols",
			Difficulty:  "hard",
			IsEndurance: 0,
			Stage:       "Sintaxis Mixta",
			TargetWPM:   18.0,
			MinAccuracy: 0.90,
		})
	}

	// Etapa 6: Expresiones Regulares y Patrones Complejos - 5 lecciones (Target: 20 WPM, 90% Acc)
	regexes := []string{
		`const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;`,
		`const phoneRegex = /^\+?[1-9]\d{1,14}$/;`,
		`const dateRegex = /^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$/;`,
		`const ipRegex = /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/;`,
		`find . -name "*.go" | xargs grep -n "SeedExercises" | sort -u`,
	}
	for i, r := range regexes {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Expresiones Regulares %d", i+1),
			Content:     r,
			Category:    "symbols",
			Difficulty:  "hard",
			IsEndurance: 0,
			Stage:       "Expresiones Regulares",
			TargetWPM:   20.0,
			MinAccuracy: 0.90,
		})
	}

	return list
}

// -------------------------------------------------------------
// CODE PROGRESSION (50 lessons)
// -------------------------------------------------------------
func generateCodeExercises(rng *rand.Rand) []TempExercise {
	var list []TempExercise

	// Etapa 1: Declaraciones Cortas (10 lecciones) - Target: 14 WPM, 88% Acc
	shortStatements := []string{
		"const total = items.reduce((acc, curr) => acc + curr, 0);",
		"val, err := db.QueryRow(\"SELECT name FROM profiles WHERE id = ?\", id)",
		"def get_user_by_id(user_id: int) -> dict:",
		"SELECT p.id, p.name, p.xp FROM profiles p WHERE p.level >= 5;",
		"const activeProfiles = computed(() => profiles().filter(p => p.streak > 0));",
		"if err != nil { return fmt.Errorf(\"failed to load data: %w\", err) }",
		"let [count, setCount] = useState<number>(0);",
		"UPDATE profiles SET streak = streak + 1 WHERE id = 1;",
		"res.status(200).json({ status: \"OK\", timestamp: Date.now() });",
		"for i := 0; i < len(exercises); i++ { fmt.Println(exercises[i].Title) }",
	}
	for i, s := range shortStatements {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Declaraciones Cortas %d", i+1),
			Content:     s,
			Category:    "code",
			Difficulty:  "easy",
			IsEndurance: 0,
			Stage:       "Declaraciones Cortas",
			TargetWPM:   14.0,
			MinAccuracy: 0.88,
		})
	}

	// Etapa 2: Estructuras de Control (10 lecciones) - Target: 18 WPM, 90% Acc
	controlStructures := []string{
		`if (status === 'active') {
    renderDashboard();
} else {
    redirectToLogin();
}`,
		`for (const key of Object.keys(metrics)) {
    if (metrics[key].errors > 5) {
        flagKeyAsWeak(key);
    }
}`,
		`switch (theme) {
case 'cyberpunk':
    applyNeonTheme();
    break;
default:
    applyGlassTheme();
}`,
		`if err != nil {
    log.Printf("error starting server: %v", err)
    os.Exit(1)
}`,
		`try {
    const data = JSON.parse(rawPayload);
    processSession(data);
} catch (err) {
    console.error("Invalid JSON:", err);
}`,
		`while (cursor < buffer.length) {
    processByte(buffer[cursor]);
    cursor++;
}`,
		`BEGIN TRANSACTION;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;`,
		`if (!profile.sound_enabled) {
    audioContext.suspend();
} else {
    audioContext.resume();
}`,
		`if (wpm >= targetWpm && accuracy >= minAccuracy) {
    unlockNextLesson();
}`,
		`router.get('/health', (req, res) => {
    res.json({ uptime: process.uptime(), status: 'UP' });
});`,
	}
	for i, c := range controlStructures {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Estructuras de Control %d", i+1),
			Content:     c,
			Category:    "code",
			Difficulty:  "medium",
			IsEndurance: 0,
			Stage:       "Estructuras de Control",
			TargetWPM:   18.0,
			MinAccuracy: 0.90,
		})
	}

	// Etapa 3: Funciones y Modelos (15 lecciones) - Target: 22 WPM, 90% Acc
	functionsModels := []string{
		`export interface Exercise {
    id: string;
    title: string;
    content: string;
    stage: string;
    target_wpm: number;
    min_accuracy: number;
}`,
		`func calculateWPM(chars int, seconds int) float64 {
    if seconds <= 0 {
        return 0.0
    }
    minutes := float64(seconds) / 60.0
    return (float64(chars) / 5.0) / minutes
}`,
		`SELECT 
    km.char,
    CAST(km.errors AS REAL) / CAST(km.attempts AS REAL) as error_rate
FROM key_metrics km
WHERE km.profile_id = 1 AND km.attempts >= 10
ORDER BY error_rate DESC LIMIT 5;`,
		`type ProfileState = {
    activeProfile: Profile | null;
    isLoading: boolean;
    sessionsCount: number;
};`,
		`func (s *Session) IsPassed(targetWPM float64, minAcc float64) bool {
    return s.WPM >= targetWPM && s.Accuracy >= minAcc
}`,
		"export class SoundService {\n    playKeyClick(switchType: string): void {\n        const audio = new Audio(\"/sounds/\" + switchType + \".mp3\");\n        audio.play();\n    }\n}",
		`SELECT e.id, e.title, MAX(s.wpm) as best_wpm
FROM exercises e
LEFT JOIN sessions s ON e.id = s.exercise_id
GROUP BY e.id
ORDER BY e.id ASC;`,
		"export function formatDuration(seconds: number): string {\n    const mins = Math.floor(seconds / 60);\n    const secs = seconds % 60;\n    return mins + \":\" + (secs < 10 ? \"0\" : \"\") + secs;\n}",
		`func NewRouter(h *handlers.Handler) *http.ServeMux {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /api/exercises", h.GetExercises)
    return mux
}`,
		`interface KeyTelemetry {
    time_ms: number;
    char: string;
    latency_ms: number;
    is_error: boolean;
}`,
		`func seedDatabase(db *sql.DB) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    return tx.Commit()
}`,
		`const debounce = (fn: Function, ms = 300) => {
    let timeoutId: ReturnType<typeof setTimeout>;
    return function (this: any, ...args: any[]) {
        clearTimeout(timeoutId);
        timeoutId = setTimeout(() => fn.apply(this, args), ms);
    };
};`,
		`SELECT category, AVG(wpm) as avg_speed, AVG(accuracy) as avg_acc
FROM sessions s
JOIN exercises e ON s.exercise_id = e.id
GROUP BY category;`,
		`export const calculateAccuracy = (correct: number, total: number): number => {
    if (total === 0) return 100;
    return Math.round((correct / total) * 100);
};`,
		`type KeyMap = {
    [code: string]: { label: string; finger: string; hand: 'left' | 'right' };
};`,
	}
	for i, f := range functionsModels {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Funciones y Modelos %d", i+1),
			Content:     f,
			Category:    "code",
			Difficulty:  "medium",
			IsEndurance: 0,
			Stage:       "Funciones y Modelos",
			TargetWPM:   22.0,
			MinAccuracy: 0.90,
		})
	}

	// Etapa 4: Código de Producción (15 lecciones) - Target: 26 WPM, 92% Acc
	for i := 0; i < 15 && i < len(codeSnippets); i++ {
		snippet := codeSnippets[i]
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Producción: %s", snippet.Title),
			Content:     snippet.Content,
			Category:    "code",
			Difficulty:  "hard",
			IsEndurance: 0,
			Stage:       "Código de Producción",
			TargetWPM:   26.0,
			MinAccuracy: 0.92,
		})
	}

	return list
}

// -------------------------------------------------------------
// NUMBERS PROGRESSION (100 lessons)
// -------------------------------------------------------------
func generateNumbersExercises(rng *rand.Rand) []TempExercise {
	var list []TempExercise

	// Etapa 1: Dígitos Básicos (25 lecciones) - Target: 15 WPM, 90% Acc
	basicNumberPatterns := []string{
		"12 34 56 78 90 12 34 56 78 90",
		"123 456 789 012 345 678 901",
		"98 76 54 32 10 98 76 54 32 10",
		"11 22 33 44 55 66 77 88 99 00",
		"10 20 30 40 50 60 70 80 90 100",
		"13 24 35 46 57 68 79 80 91 02",
		"101 202 303 404 505 606 707 808",
		"1212 3434 5656 7878 9090",
		"54321 12345 67890 09876",
		"2 4 6 8 10 12 14 16 18 20",
		"1 3 5 7 9 11 13 15 17 19 21",
		"100 200 300 400 500 600 700 800",
		"999 888 777 666 555 444 333",
		"12 123 1234 12345 123456",
		"654321 54321 4321 321 21 1",
		"15 25 35 45 55 65 75 85 95",
		"10 100 1000 10000 100000",
		"24 48 72 96 120 144 168 192",
		"50 150 250 350 450 550 650",
		"123 321 456 654 789 987",
		"111 222 333 444 555 666 777",
		"14 28 42 56 70 84 98 112",
		"30 60 90 120 150 180 210",
		"16 32 64 128 256 512 1024",
		"1024 2048 4096 8192 16384",
	}
	for i, p := range basicNumberPatterns {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Dígitos Básicos %d", i+1),
			Content:     p,
			Category:    "numbers",
			Difficulty:  "easy",
			IsEndurance: 0,
			Stage:       "Dígitos Básicos",
			TargetWPM:   15.0,
			MinAccuracy: 0.90,
		})
	}

	// Etapa 2: Formatos Numéricos Comunes (35 lecciones) - Target: 20 WPM, 92% Acc
	contextNumbers := []string{
		"El año 2026 marca un nuevo hito en el desarrollo de software.",
		"Niko entrena con una barra de 30kg y mancuernas ajustables de 15kg.",
		"En casa somos una familia de 6 personas: 2 adultos y 4 hijos.",
		"Lograr 100% de precisión a 45 WPM requiere práctica constante.",
		"La guitarra LTD SC-207 cuenta con 7 cuerdas y escala de 25.5 pulgadas.",
		"El amplificador Line 6 Spider entrega 150W de potencia estéreo.",
		"El perfil alcanzó el nivel 28 con más de 15,000 puntos de XP.",
		"Completó las 50 lecciones con una racha activa de 14 días seguidos.",
		"La temperatura ambiente es de 22°C con una humedad relativa del 45%.",
		"La sesión duró 180 segundos con un total de 420 caracteres escritos.",
		"Nació el 23 de noviembre de 1984 en la ciudad de Santiago.",
		"El servidor web procesa 2,500 peticiones por segundo sin latencia.",
		"El disco SSD tiene una capacidad de 512GB y 16GB de memoria RAM.",
		"El precio total fue de $45,990 pesos con un descuento del 15%.",
		"El récord clásico es de 68 WPM con un 98.5% de precisión global.",
		"La reunión está programada para las 14:30 horas del viernes 12.",
		"El microservicio responde en 15ms con un consumo de 48MB de memoria.",
		"El porcentaje de error disminuyó del 8.5% al 1.2% en 3 semanas.",
		"Se registraron 1,200 visitas diarias y 350 usuarios concurrentes.",
		"El archivo comprimido pesa 24.8MB y contiene 150 elementos.",
		"Entrenar 3 a 4 veces por semana previene dolores musculares a los 40 años.",
		"La velocidad promedio aumentó 12 WPM tras completar las lecciones.",
		"El vehículo recorrió 120km en 75 minutos por la autopista central.",
		"El procesador Intel Core i9 cuenta con 8 núcleos y 16 hilos a 2.4GHz.",
		"El monitor de 27 pulgadas tiene una resolución de 2560x1440 píxeles a 144Hz.",
		"En el torneo participaron 64 jugadores divididos en 8 grupos de 8.",
		"La meta diaria de pasos es de 10,000 con 45 minutos de caminata.",
		"El proyecto cuenta con 40 pruebas unitarias y un 95% de cobertura.",
		"El tanque de combustible tiene 55 litros y rinde 14.5km por litro.",
		"El plazo de entrega es de 48 horas hábiles a partir de las 09:00.",
		"La frecuencia de muestreo es de 44.1kHz con una resolución de 24 bits.",
		"Se aplicó un aumento del 3.5% en la tasa de interés anual.",
		"El puente mide 450 metros de longitud y 24 metros de ancho.",
		"El teclado tiene 87 teclas mecánicas con iluminación RGB de 16.8M colores.",
		"La velocidad de descarga alcanzó los 300Mbps con 50Mbps de subida.",
	}
	for i, c := range contextNumbers {
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Formatos Numéricos %d", i+1),
			Content:     c,
			Category:    "numbers",
			Difficulty:  "medium",
			IsEndurance: 0,
			Stage:       "Formatos Numéricos",
			TargetWPM:   20.0,
			MinAccuracy: 0.92,
		})
	}

	// Etapa 3: Secuencias Técnicas e IPs (40 lecciones) - Target: 25 WPM, 92% Acc
	technicalNumbers := []string{
		"192.168.1.1, 192.168.1.100, 192.168.1.254",
		"127.0.0.1:8080, 127.0.0.1:3000, 127.0.0.1:5432",
		"10.0.0.1:27017, 10.0.0.2:6379, 10.0.0.5:9092",
		"8.8.8.8, 8.8.4.4, 1.1.1.1, 1.0.0.1",
		"+56 9 9876 5432, +56 2 2345 6789",
		"2026-01-15, 2026-06-24, 2026-12-31",
		"3.14159265, 2.71828182, 1.41421356",
		"0.0.0.0/0, 192.168.0.0/24, 10.0.0.0/16",
		"255.255.255.0, 255.255.0.0, 255.0.0.0",
		"HTTP 200 OK, HTTP 404 Not Found, HTTP 500 Error",
		"Port: 22 (SSH), Port: 80 (HTTP), Port: 443 (HTTPS)",
		"UUID: 550e8400-e29b-41d4-a716-446655440000",
		"Git hash: a1b2c3d4e5f60718293a4b5c6d7e8f90",
		"v1.0.0, v1.2.4, v2.0.1-beta.3, v3.5.0-rc.1",
		"RGB(255, 128, 0), RGBA(0, 240, 255, 0.85)",
		"Lat: -33.6067, Long: -70.5756, Alt: 650m",
		"192.168.100.1:8443, 172.16.0.1:9000",
		"3840x2160 @ 60Hz, 1920x1080 @ 144Hz",
		"MAC: 00:1A:2B:3C:4D:5E, IP: 192.168.1.50",
		"MD5: e4d909c290d0fb1ca068ffaddf22cbd0",
	}
	for i := 0; i < 40; i++ {
		tn := technicalNumbers[i%len(technicalNumbers)]
		list = append(list, TempExercise{
			Title:       fmt.Sprintf("Secuencias Técnicas %d", i+1),
			Content:     tn,
			Category:    "numbers",
			Difficulty:  "hard",
			IsEndurance: 0,
			Stage:       "Secuencias Técnicas",
			TargetWPM:   25.0,
			MinAccuracy: 0.92,
		})
	}

	return list
}

var (
	spanishSentences = []string{
		"El entrenamiento de hipertrofia con mancuernas ajustables y barra de 30kg ayuda a ganar fuerza en casa.",
		"Escribir código limpio en Go requiere disciplina y paciencia para estructurar de manera óptima los paquetes.",
		"Tocar un solo rápido de Megadeth con la guitarra LTD SC-207 de siete cuerdas mejora la coordinación motora.",
		"En una familia numerosa con niños neurodivergentes, la paciencia y el apoyo emocional son los pilares fundamentales.",
		"El despliegue de microservicios con Docker simplifica enormemente la administración y el escalado de servidores.",
		"El death metal melódico de In Flames combina ritmos agresivos con guitarras armonizadas sumamente complejas.",
		"La constancia en los entrenamientos de fuerza, unas tres o cuatro veces por semana, es clave para evitar lesiones.",
		"El desarrollo de frontend en Angular permite crear interfaces dinámicas estructuradas mediante componentes y señales.",
		"PostgreSQL es una base de datos excelente, robusta y confiable para gestionar proyectos complejos de backend en Go.",
		"En casa nos organizamos usando un sistema basado en BuJo y GTD para coordinar las tareas familiares del día a día.",
		"Amon Amarth utiliza temáticas vikingas y ritmos pesados para energizar al máximo sus composiciones en vivo.",
		"El uso de NATS para mensajería distribuida facilita la comunicación asíncrona y en tiempo real entre servicios.",
		"Administrar las tareas del hogar requiere un equilibrio constante entre una disciplina clara y contención afectiva.",
		"Aprender mecanografía táctil es muy similar a practicar escalas en la guitarra, ya que exige repetición deliberada.",
		"El descanso adecuado y un volumen de entrenamiento controlado son fundamentales para la ganancia de masa muscular.",
		"Un desarrollador senior escribe pruebas unitarias robustas para garantizar la estabilidad y escalabilidad del software.",
		"Niko utiliza su Keychron K8 con interruptores mecánicos para escribir código de forma cómoda y sumamente veloz.",
		"La música metal es el mejor acompañamiento para entrenar pesado con barra libre y mancuernas ajustables en casa.",
		"Nuestros hijos Pablo, Feña, Sofi y Luciano tienen personalidades únicas que llenan la casa de alegría y dinamismo.",
		"Integrar pipelines de integración continua con Docker facilita la entrega de software sin fricciones al usuario final.",
		"Megadeth destaca en el thrash metal por sus riffs intrincados y los solos ultra veloces de Dave Mustaine.",
		"La neurodivergencia en la familia nos enseña a ser flexibles y a valorar cada pequeño avance en el aprendizaje.",
		"Configurar variables de entorno y secretos de forma segura previene vulnerabilidades críticas en la infraestructura.",
		"El entrenamiento de fuerza máxima requiere levantar cargas pesadas con una técnica impecable para evitar accidentes.",
		"La guitarra LTD SC-207 cuenta con un mástil cómodo y siete cuerdas que amplían el rango de tonos graves.",
		"Escribir scripts de automatización en Python ahorra horas de trabajo manual en tareas repetitivas de backend.",
		"Una buena arquitectura de software separa las responsabilidades de la base de datos de las reglas de negocio.",
		"En La Tropa nos encanta compartir momentos juntos, desde escuchar música pesada hasta jugar juegos de mesa.",
		"Levantar peso muerto con una barra de 30kg fortalece la cadena posterior y mejora la postura corporal diaria.",
		"Metallica popularizó los tempos rápidos y las estructuras complejas en sus primeros álbumes de estudio de thrash.",
		"Gestionar el TDAH y el autismo con empatía y rutinas claras ayuda a los niños a sentirse seguros en el hogar.",
		"Las APIs REST escritas en Go destacan por su bajo consumo de memoria y su altísimo rendimiento bajo carga.",
		"Optimizar los índices en PostgreSQL puede acelerar drásticamente las consultas de reportes en el backend.",
		"Usar switch mecánicos de tipo lineal o táctil depende de las preferencias personales al escribir texto o código.",
		"El metal de Gotemburgo influyó a toda una generación de músicos con sus características melodías de guitarra.",
		"La hipertrofia se logra combinando tensión mecánica, estrés metabólico y un adecuado aporte de proteínas diarias.",
		"Los componentes funcionales y el manejo de estados reactivos simplifican el mantenimiento de las vistas de usuario.",
		"Coordinar agendas de seis personas en una familia ensamblada requiere una comunicación abierta y honesta.",
		"La automatización de pruebas con Selenium garantiza que los flujos críticos de la aplicación sigan operativos.",
		"Utilizar Docker Compose facilita levantar todo el entorno de base de datos y backend con un solo comando.",
		"James Hetfield es reconocido por su excelente técnica de púa hacia abajo al ejecutar riffs de guitarra rítmica.",
		"Encontrar el equilibrio entre disciplina deportiva y desarrollo profesional previene el desgaste físico y mental.",
		"Luciano tiene un pensamiento sumamente lógico y detallista que le permite resolver rompecabezas muy complejos.",
	}

	englishSentences = []string{
		"Hypertrophy training with adjustable dumbbells and a 30kg barbell promotes muscle growth at home.",
		"Writing clean idiomatic Go requires discipline and patience to structure your packages properly.",
		"Playing a fast Megadeth guitar solo on a seven-string LTD SC-207 improves motor finger coordination.",
		"In a large family with neurodivergent children, patience and emotional containment are essential.",
		"Deploying microservices with Docker Compose simplifies server orchestration and rapid scalability.",
		"In Flames and melodic death metal blend aggressive tempos with intricate twin-guitar harmonies.",
		"Lifting weights three to four times a week consistently builds long-term physical resilience.",
		"Angular signals provide a reactive and declarative model to handle client-side state smoothly.",
		"PostgreSQL offers rock-solid reliability for demanding backend enterprise applications.",
		"Our home routine relies on Bullet Journal and GTD methodologies to organize daily commitments.",
		"Amon Amarth channels heavy Viking themes and thunderous rhythms in their energetic concerts.",
		"Distributed messaging with NATS enables low-latency communication across decoupled microservices.",
		"Balancing athletic discipline with remote work prevents burnout and sustains mental sharpness.",
		"Touch typing practice is similar to rehearsing guitar arpeggios through deliberate repetition.",
		"Adequate sleep and controlled training volume are fundamental pillars for muscle recovery.",
		"Senior software engineers write comprehensive unit tests to prevent costly regression bugs.",
		"Niko relies on his Keychron K8 mechanical keyboard to write software fast with tactile clarity.",
		"Heavy metal music provides an empowering soundtrack for focused barbell workouts in the garage.",
		"Our kids Pablo, Feña, Sofi, and Luciano bring immense energy and vibrant joy to our household.",
		"Continuous integration pipelines with automated Docker tests streamline software deployments.",
		"Dave Mustaine and Megadeth pioneered aggressive thrash metal riffs with razor-sharp precision.",
		"Embracing neurodivergence teaches us flexibility, empathy, and joy in celebrating small victories.",
		"Securing environment variables and secrets prevents severe vulnerabilities across cloud services.",
		"Lifting heavy deadlifts with proper posture strengthens the posterior chain and core stability.",
		"The LTD SC-207 guitar features a smooth neck profile and extended range for crushing low notes.",
		"Automation scripts written in Python save hours of manual testing in backend data migrations.",
		"Clean architecture separates database infrastructure concerns from core business business logic.",
		"La Tropa enjoys sharing loud metal concerts, homemade dinners, and strategic board games together.",
		"Consistent deadlift workouts build enduring spinal support and functional whole-body power.",
		"Metallica popularized rapid thrash rhythms and complex song structures in their early masterworks.",
		"Providing structured routines helps children with ADHD and autism navigate daily transitions safely.",
		"Go microservices excel under heavy concurrent traffic with minimal memory overhead and fast boot.",
		"Tuning composite indexes in PostgreSQL accelerates complex aggregation queries substantially.",
		"Choosing between tactile and linear mechanical switches is a matter of personal acoustic taste.",
		"Gothenburg melodic death metal inspired generations of guitar players with soaring harmonized leads.",
		"Muscle hypertrophy requires mechanical tension, progressive overload, and sufficient daily protein.",
		"Modular reactive components keep single-page web applications maintainable as features expand.",
		"Harmonizing schedules for a family of six demands transparent communication and mutual patience.",
		"End-to-end automated testing with Selenium ensures critical user checkout flows never fail.",
		"Docker Compose makes spinning up local databases and backend mock servers effortless and quick.",
		"Down-picking speed and rhythmic tightness are hallmarks of world-class metal rhythm guitarists.",
		"Maintaining a healthy balance between coding and barbell training fosters lifelong longevity.",
		"Luciano exhibits impressive logical problem-solving abilities when tackling intricate visual puzzles.",
	}

	codeSnippets = []struct {
		Title      string
		Language   string
		Difficulty string
		Content    string
	}{
		{
			Title: "Go: Channel Worker Pool",
			Language: "Go",
			Difficulty: "medium",
			Content: `func worker(id int, jobs <-chan int, results chan<- int) {
    for j := range jobs {
        results <- j * 2
    }
}`,
		},
		{
			Title: "TypeScript: Observable Store",
			Language: "TypeScript",
			Difficulty: "medium",
			Content: `export class ProfileStore {
    private readonly state = signal<ProfileState>(initialState);
    readonly profile = computed(() => this.state().activeProfile);
}`,
		},
		{
			Title: "SQL: Aggregate Reporting",
			Language: "SQL",
			Difficulty: "medium",
			Content: `SELECT p.name, COUNT(s.id) AS total_sessions, AVG(s.wpm) AS avg_speed
FROM profiles p
JOIN sessions s ON p.id = s.profile_id
GROUP BY p.id
HAVING AVG(s.accuracy) >= 0.90;`,
		},
		{
			Title: "Go: HTTP JSON Handler",
			Language: "Go",
			Difficulty: "medium",
			Content: `func HealthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "UP"})
}`,
		},
		{
			Title: "Python: Data Pipeline Filter",
			Language: "Python",
			Difficulty: "medium",
			Content: `def filter_valid_sessions(records: list[dict]) -> list[dict]:
    return [r for r in records if r.get("accuracy", 0) >= 0.90 and r.get("wpm", 0) >= 20]`,
		},
		{
			Title: "SQL: Window Function Rank",
			Language: "SQL",
			Difficulty: "hard",
			Content: `WITH RankedSessions AS (
    SELECT profile_id, wpm, accuracy,
           ROW_NUMBER() OVER (PARTITION BY profile_id ORDER BY wpm DESC) as rank
    FROM sessions
)
SELECT profile_id, wpm as best_wpm FROM RankedSessions WHERE rank = 1;`,
		},
		{
			Title: "TypeScript: Generic Result Type",
			Language: "TypeScript",
			Difficulty: "hard",
			Content: `export type Result<T, E = Error> = 
    | { ok: true; value: T }
    | { ok: false; error: E };`,
		},
		{
			Title: "Go: Mutex Protected Cache",
			Language: "Go",
			Difficulty: "hard",
			Content: `type SafeCache struct {
    mu    sync.RWMutex
    items map[string]interface{}
}
func (c *SafeCache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.items[key]
    return val, ok
}`,
		},
	}
)
