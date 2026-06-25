package db

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"
)

type TempExercise struct {
	Title       string
	Content     string
	Category    string
	Difficulty  string
	IsEndurance int
	SourceIDs   []int // Índices de origen del pool para control de dispersión
}

// SeedExercises checks if exercises table is empty or stale, and seeds unique exercises progressively
func SeedExercises(db *sql.DB) {
	// Detect and clean old base-0 exercises (like 'spa_0' or 'eng_0')
	var oldExists bool
	_ = db.QueryRow("SELECT EXISTS(SELECT 1 FROM exercises WHERE id = 'spa_0' OR id = 'eng_0')").Scan(&oldExists)
	if oldExists {
		log.Println("Old base-0 exercises detected. Purging exercises table for fresh base-1 seeds...")
		_, err := db.Exec("DELETE FROM exercises")
		if err != nil {
			log.Printf("Error purging old exercises: %v", err)
		}
	}

	// Detect and clean old repetitive exercises
	var hasOldTemplate bool
	_ = db.QueryRow("SELECT EXISTS(SELECT 1 FROM exercises WHERE content LIKE '%permite ganar fuerza e LTD%')").Scan(&hasOldTemplate)
	if hasOldTemplate {
		log.Println("Old repetitive template detected. Purging exercises table for new rich semantic seeds...")
		_, err := db.Exec("DELETE FROM exercises")
		if err != nil {
			log.Printf("Error purging old exercises: %v", err)
		}
	}

	// Detect and clean old excessive or undispersed seeds
	var codeCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM exercises WHERE category = 'code'").Scan(&codeCount)
	if codeCount > 25 {
		log.Println("Old excessive or undispersed seeds detected. Purging exercises table...")
		_, err := db.Exec("DELETE FROM exercises")
		if err != nil {
			log.Printf("Error purging exercises: %v", err)
		}
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&count)
	if err != nil {
		log.Fatalf("Error checking exercise count: %v", err)
	}

	if count >= 700 {
		log.Printf("Exercises already seeded (%d exercises in DB)", count)
		return
	}

	log.Println("Generating 700 unique exercises in memory...")
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Seed target allocations
	totalCode := 25
	totalSpanish := 250
	totalEnglish := 250
	totalNumbers := 100
	totalSymbols := 75

	// Generate into memory first
	var tempExercises []TempExercise
	tempExercises = append(tempExercises, generateCodeExercises(totalCode, rng)...)
	tempExercises = append(tempExercises, generateSpanishExercises(totalSpanish, rng)...)
	tempExercises = append(tempExercises, generateEnglishExercises(totalEnglish, rng)...)
	tempExercises = append(tempExercises, generateNumbersExercises(totalNumbers, rng)...)
	tempExercises = append(tempExercises, generateSymbolsExercises(totalSymbols, rng)...)

	// Group by category to sort individually
	categorized := make(map[string][]TempExercise)
	for _, ex := range tempExercises {
		categorized[ex.Category] = append(categorized[ex.Category], ex)
	}

	// Helper to calculate progressive sorting scores
	scoreDiff := func(e TempExercise) int {
		if e.Difficulty == "easy" {
			return 1
		}
		if e.Difficulty == "medium" {
			return 2
		}
		// "hard"
		if e.IsEndurance == 1 {
			return 4
		}
		return 3
	}

	log.Println("Sorting, dispersing, and ordering exercises progressively...")
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}

	stmt, err := tx.Prepare("INSERT INTO exercises (id, title, content, category, difficulty, is_endurance) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatalf("Error preparing statement: %v", err)
	}
	defer stmt.Close()

	// Sort, disperse, and insert each category progressively from 1 to N
	for cat, list := range categorized {
		sort.Slice(list, func(i, j int) bool {
			scoreI := scoreDiff(list[i])
			scoreJ := scoreDiff(list[j])
			if scoreI != scoreJ {
				return scoreI < scoreJ
			}
			return len(list[i].Content) < len(list[j].Content)
		})

		// Disperse to prevent adjacent exercises from sharing source items
		list = disperseExercises(list)

		// Insert with progressive IDs and Titles
		for index, ex := range list {
			var id, finalTitle string
			seq := index + 1 // Base 1 numbering

			switch cat {
			case "code":
				id = fmt.Sprintf("code_%d", seq)
				finalTitle = fmt.Sprintf("%s %d", ex.Title, seq)
			case "spanish":
				id = fmt.Sprintf("spa_%d", seq)
				finalTitle = fmt.Sprintf("Español: Lección %d", seq)
			case "english":
				id = fmt.Sprintf("eng_%d", seq)
				finalTitle = fmt.Sprintf("English: Lesson %d", seq)
			case "numbers":
				id = fmt.Sprintf("num_%d", seq)
				finalTitle = fmt.Sprintf("Números: Lección %d", seq)
			case "symbols":
				id = fmt.Sprintf("sym_%d", seq)
				finalTitle = fmt.Sprintf("Símbolos: Lección %d", seq)
			default:
				id = fmt.Sprintf("ex_%s_%d", cat, seq)
				finalTitle = fmt.Sprintf("Lección %d", seq)
			}

			_, err = stmt.Exec(id, finalTitle, ex.Content, ex.Category, ex.Difficulty, ex.IsEndurance)
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
	log.Printf("Successfully seeded %d progressive exercises (Base-1) in database", newCount)
}

// disperseExercises sweeps the list and makes sure consecutive exercises do not share any source item index
func disperseExercises(list []TempExercise) []TempExercise {
	n := len(list)
	for i := 1; i < n; i++ {
		if sharesSource(list[i], list[i-1]) {
			// Find the next element j > i that doesn't share elements with list[i-1]
			found := false
			for j := i + 1; j < n; j++ {
				if !sharesSource(list[j], list[i-1]) {
					list[i], list[j] = list[j], list[i]
					found = true
					break
				}
			}
			_ = found
		}
	}
	return list
}
func sharesSource(a, b TempExercise) bool {
	for _, sa := range a.SourceIDs {
		for _, sb := range b.SourceIDs {
			if sa == sb {
				return true
			}
		}
	}
	return false
}

// pools de datos enriquecidos para generación dinámica de lecciones
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
		"Configurar middleware en nuestro servidor Go nos ayuda a registrar métricas y latencias de cada petición HTTP.",
		"El uso de variables reactivas en Angular previene renderizados innecesarios y optimiza la velocidad del frontend.",
		"Hacer ejercicio en casa ofrece una gran flexibilidad horaria para complementar el trabajo remoto diario.",
		"La Samick de seis cuerdas es una guitarra clásica excelente para practicar acordes y progresiones básicas.",
		"Documentar las APIs con Swagger permite que otros desarrolladores integren sus servicios de forma autónoma.",
		"Un pilar de estabilidad en la familia ayuda a regular las emociones de todos en momentos de alta tensión.",
		"Aprender a escribir sin mirar el teclado libera capacidad mental para enfocarse únicamente en la lógica del texto.",
		"Los microservicios comunicados por NATS son altamente tolerantes a fallos y fáciles de escalar horizontalmente.",
		"El volumen de entrenamiento semanal se debe ajustar gradualmente para asegurar una recuperación muscular completa.",
		"El press de banca con barra de 30kg estimula el crecimiento de los pectorales y mejora la fuerza de empuje.",
		"Dave Mustaine compuso riffs legendarios de Megadeth que requieren gran velocidad y precisión en la púa.",
		"La Tropa siempre viaja unida, haciendo que cada salida familiar sea una aventura caótica pero hermosa.",
		"Implementar comunicación asíncrona con NATS en Go permite construir servicios sumamente desacoplados.",
		"Configurar la LTD SC-207 de siete cuerdas en afinación estándar de si menor proporciona tonos oscuros ideales para el metal.",
		"El sistema de productividad GTD combinado con Bullet Journal ayuda a mantener el foco y reducir la sobrecarga mental.",
		"Un teclado Keychron K8 con interruptores táctiles ofrece una respuesta excelente para largas jornadas de programación.",
		"Luciano se enfoca intensamente en sus intereses lógicos y disfruta analizando patrones complejos en la computadora.",
		"Hacer sentadillas con barra libre en casa exige una excelente alineación de la espalda para evitar lesiones lumbares.",
		"Docker permite empaquetar aplicaciones Go y Node con todas sus dependencias en contenedores ligeros.",
		"Los solos armonizados de In Flames definieron el sonido característico del death metal melódico de Gotemburgo.",
		"Feña muestra una gran sensibilidad y empatía, siendo siempre el pilar de apoyo emocional entre sus hermanos.",
		"Realizar pruebas de carga con JMeter asegura que las APIs de backend toleren miles de peticiones simultáneas.",
		"Las mancuernas ajustables son la herramienta perfecta para entrenar hipertrofia en casa optimizando el espacio.",
		"PostgreSQL maneja transacciones ACID de forma segura, garantizando la integridad de datos críticos del negocio.",
		"El ratón Logitech M720 permite alternar rápidamente entre la laptop del trabajo y la computadora personal.",
		"Amon Amarth ofrece un espectáculo en vivo impresionante, recreando barcos vikingos sobre el escenario.",
		"Coordinar los apoyos terapéuticos y rutinas estructuradas en casa es clave para el bienestar de Luciano.",
		"El desarrollo frontend con React y Angular se beneficia enormemente del uso de estados reactivos y señales.",
		"El volumen de entrenamiento de hipertrofia debe calcularse en base a series efectivas semanales por grupo muscular.",
		"Selenium automatiza flujos complejos en el navegador, reduciendo drásticamente el tiempo de pruebas de regresión.",
		"Practicar escalas pentatónicas a tempo lento con metrónomo construye una sólida memoria muscular en los dedos.",
		"La Tropa se reúne los fines de semana para disfrutar de una parrillada y escuchar buen thrash metal clásico.",
		"El compilador de Go genera binarios estáticos extremadamente rápidos y eficientes para entornos en la nube.",
		"El entrenamiento de fuerza fortalece los tendones y mejora la densidad ósea a lo largo de los años.",
		"Automatizar tareas repetitivas de infraestructura con scripts de Python ahorra valiosas horas de trabajo manual.",
		"Mantener rutinas estables y claras ayuda a regular la ansiedad en perfiles neurodivergentes.",
		"El amplificador Line 6 Spider IV 150 ofrece una gran variedad de efectos y modelado de amplificadores clásicos.",
		"Utilizar Postman para probar endpoints facilita la documentación y el diseño de APIs REST robustas.",
		"Configurar pipelines de CI/CD facilita el despliegue automático en servicios de nube como AWS.",
	}

	englishSentences = []string{
		"Regular strength training at home with adjustable dumbbells builds solid muscle and physical endurance.",
		"Writing clean Go code requires deep focus on package modularity and proper concurrency design patterns.",
		"Playing a fast Megadeth guitar solo on a seven-string LTD requires slow practice and muscle memory.",
		"In a large family with neurodivergent kids, patience and emotional support are key to daily harmony.",
		"Deploying microservices with Docker simplifies server management and scaling operations significantly.",
		"Melodic death metal from Gothenburg combines aggressive tempos with highly complex dual guitar leads.",
		"Consistent hypertrophy training three or four times a week is essential for preventing joint injuries.",
		"Developing client applications with Angular allows structuring dynamic interfaces using reactive signals.",
		"PostgreSQL is a robust and reliable database management system for handling complex backend logic in Go.",
		"We organize our family tasks using a hybrid system based on Bullet Journal and Getting Things Done.",
		"Amon Amarth utilizes Norse mythology themes and heavy riffs to deliver powerful live performances.",
		"Using NATS for distributed messaging facilitates asynchronous and real-time communication between services.",
		"Managing household chores in a big family requires a balance between structure and emotional support.",
		"Learning touch typing is very similar to practicing scales on a guitar, demanding deliberate repetition.",
		"Adequate sleep and controlled training volume are crucial for achieving optimal muscle hypertrophy.",
		"A senior developer writes comprehensive unit tests to ensure long-term stability and code quality.",
		"Writing automation scripts in Python saves hours of manual labor in repetitive backend testing tasks.",
		"Good software architecture decouples database persistence from the core application business rules.",
		"Our children Pablo, Feña, Sofi, and Luciano have unique personalities that fill our home with joy.",
		"Integrating continuous integration pipelines with Docker ensures smooth deployments to staging servers.",
		"Megadeth stands out in thrash metal history due to complex rhythm patterns and fast lead guitar work.",
		"Supporting neurodivergent learners at home teaches us to be flexible and appreciate small daily victories.",
		"Configuring environment variables securely prevents critical vulnerability leaks in cloud infrastructure.",
		"Maximum strength training demands lifting heavy weights with perfect technique to avoid back injuries.",
		"The LTD seven-string guitar features a comfortable neck profile and extends the range of lower tones.",
		"Levisting weights at home offers great scheduling flexibility to balance with remote software work.",
		"The Samick electric guitar is a classic instrument that serves well for practicing scales and chords.",
		"Documenting REST APIs with OpenAPI allows team members to integrate frontend components seamlessly.",
		"A stable family environment serves as a safe harbor during stressful moments or emotional crises.",
		"Touch typing without looking at the keyboard allows you to focus completely on the logic of your code.",
		"Microservices communicating via NATS are highly resilient and easy to scale up horizontally.",
		"Weekly training volume should be adjusted gradually to ensure full muscular and nervous recovery.",
		"Developing modular frontend features in Angular keeps the codebase clean and easy to test over time.",
		"James Hetfield is widely recognized for his precision and power in downpicking complex rhythm guitar riffs.",
		"Docker Compose makes it extremely easy to spin up databases and backend services with a single script.",
		"Luciano has a highly logical mind and attention to detail that helps him solve difficult cognitive tasks.",
		"Configuring custom middleware in Go projects logs execution latency and traces incoming requests.",
		"Hypertrophy is stimulated by combining mechanical tension, metabolic stress, and proper nutrition.",
		"Patience and logical routines help neurodivergent children feel secure and focused at home.",
		"Dave Mustaine is renowned for his technical guitar playing and fast speed in thrash metal songs.",
		"Managing a family of six demands clear boundaries, open communication, and constant organization.",
		"Using mechanical switches with the right actuation force improves typing comfort and key feedback.",
		"Swedish metal bands like In Flames influenced modern metalcore with their melodic twin guitar riffs.",
		"Automating integration tests with Selenium ensures critical user pathways remain functional.",
		"PostgreSQL query performance can be boosted by creating partial indexes and optimizing joins.",
		"Go microservices are lightweight, consume very little RAM, and process requests with minimal latency.",
		"Lifting a thirty kilogram barbell on deadlifts strengthens your back and stabilizes your core posture.",
		"Keeping active streaks on productive habits helps build consistency and leads to personal mastery.",
		"Adjusting dumbbells between sets allows you to target different muscle groups effectively at home.",
		"A senior software engineer always reviews pull requests with constructive feedback to help junior developers.",
		"Bench press with a thirty kilogram barbell builds chest strength and shoulder stability over time.",
		"Dave Mustaine created iconic Megadeth riffs that demand exceptional downpicking speed and precision.",
		"La Tropa always travels together, turning every family trip into a beautifully chaotic adventure.",
		"Implementing asynchronous messaging with NATS in Go helps build highly decoupled microservices.",
		"Tuning the seven-string LTD guitar to standard B brings out the deep, heavy tones ideal for modern metal.",
		"A productivity workflow combining GTD and Bullet Journal reduces mental load and keeps tasks organized.",
		"A Keychron K8 keyboard with tactile switches provides excellent feedback for long programming sessions.",
		"Luciano excels at logical tasks and loves searching for detailed patterns on his computer screen.",
		"Performing free-weight squats at home requires strict back form to prevent lower spine injuries.",
		"Docker simplifies packaging Go and Node applications into lightweight, self-contained containers.",
		"Harmonized guitar solos by In Flames shaped the melodic death metal sound of Gothenburg.",
		"Feña displays great empathy and serves as a supportive anchor for his brothers and sisters.",
		"Running load tests with JMeter ensures backend services can handle thousands of concurrent requests.",
		"Adjustable dumbbells are the perfect solution for training hypertrophy at home with limited space.",
		"PostgreSQL handles transactional data safely, ensuring data integrity for production environments.",
		"The Logitech M720 mouse allows switching seamlessly between my work laptop and personal desktop.",
		"Amon Amarth delivers epic live shows, often featuring massive Viking ships on the stage.",
		"Coordinating therapy support and stable routines is crucial for Luciano's development and peace.",
		"Developing frontend features with React and Angular benefits from reactive state management tools.",
		"Weekly training volume should be measured in effective sets per muscle group to optimize growth.",
		"Selenium automates complex user flows in browsers, reducing regression testing time significantly.",
		"Practicing scales slowly with a metronome builds reliable finger muscle memory on the guitar.",
		"La Tropa gathers on weekends to enjoy barbecues and listen to classic thrash metal bands.",
		"The Go compiler produces fast, static binaries that run efficiently in cloud environments.",
		"Strength training increases tendon resilience and improves bone density over the years.",
		"Writing automation scripts in Python saves hours of manual labor in operations tasks.",
		"Maintaining clear and predictable schedules helps reduce anxiety in neurodivergent family members.",
		"The Line 6 Spider IV amplifier provides a wide range of presets for custom metal tones.",
		"Using Postman to test endpoints simplifies API design and speeds up backend development.",
		"Setting up CI/CD pipelines ensures automatic deployment of applications to cloud systems.",
	}

	codeSnippets = []CodeSnippet{
		// Go
		{
			Title: "Go: HTTP Handler",
			Language: "Golang",
			Difficulty: "easy",
			Content: `func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}
	profile, err := h.service.GetProfileByID(r.Context(), id)
	if err != nil {
		http.Error(w, "profile not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}`,
		},
		{
			Title: "Go: Channel Sync",
			Language: "Golang",
			Difficulty: "medium",
			Content: `func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		fmt.Printf("worker %d started job %d\n", id, j)
		time.Sleep(time.Millisecond * 100)
		results <- j * 2
	}
}

func main() {
	jobs := make(chan int, 100)
	results := make(chan int, 100)
	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}
	for j := 1; j <= 5; j++ {
		jobs <- j
	}
	close(jobs)
}`,
		},
		{
			Title: "Go: Context Timeout",
			Language: "Golang",
			Difficulty: "hard",
			Content: `func FetchDataWithTimeout(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}`,
		},
		{
			Title: "Go: Concurrencia WaitGroup",
			Language: "Golang",
			Difficulty: "medium",
			Content: `func main() {
	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Goroutine %d executing\n", id)
		}(i)
	}
	wg.Wait()
}`,
		},
		{
			Title: "Go: Mutex Locks",
			Language: "Golang",
			Difficulty: "hard",
			Content: `type SafeCounter struct {
	mu sync.Mutex
	v  map[string]int
}

func (c *SafeCounter) Inc(key string) {
	c.mu.Lock()
	c.v[key]++
	c.mu.Unlock()
}`,
		},

		// Angular / TS
		{
			Title: "Angular: Component Signals",
			Language: "Angular",
			Difficulty: "easy",
			Content: `@Component({
  selector: 'app-counter',
  standalone: true,
  template: '<button (click)="increment()">Clicked {{ count() }} times</button>'
})
export class CounterComponent {
  readonly count = signal(0);

  increment(): void {
    this.count.update(val => val + 1);
  }
}`,
		},
		{
			Title: "Angular: Computed States",
			Language: "Angular",
			Difficulty: "medium",
			Content: `@Component({
  selector: 'app-user-profile',
  standalone: true,
  imports: [CommonModule],
  template: '<div>{{ fullName() }} - XP: {{ xp() }} (Level {{ level() }})</div>'
})
export class UserProfileComponent {
  readonly firstName = signal('Niko');
  readonly lastName = signal('Parrado');
  readonly xp = signal(1500);

  readonly fullName = computed(() => this.firstName() + ' ' + this.lastName());
  readonly level = computed(() => Math.floor(this.xp() / 1000) + 1);
}`,
		},
		{
			Title: "Angular: API Injection",
			Language: "Angular",
			Difficulty: "hard",
			Content: `@Injectable({ providedIn: 'root' })
export class ProfileService {
  private http = inject(HttpClient);
  private apiUrl = 'https://api.latropa.type/profiles';

  getProfile(id: number): Observable<Profile> {
    return this.http.get<Profile>(this.apiUrl + '/' + id).pipe(
      map(data => ({ ...data, loadedAt: new Date() })),
      catchError(err => {
        console.error('Error fetching profile', err);
        return throwError(() => new Error('Profile loading failed'));
      })
    );
  }
}`,
		},
		{
			Title: "Angular: Input Binding",
			Language: "Angular",
			Difficulty: "easy",
			Content: `@Component({
  selector: 'app-user',
  standalone: true,
  template: '<h2>Welcome {{ name() }}</h2>'
})
export class UserComponent {
  readonly name = input.required<string>();
}`,
		},

		// React
		{
			Title: "React: Fetch Custom Hook",
			Language: "React",
			Difficulty: "medium",
			Content: `export function useFetch(url) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(url)
      .then(res => res.json())
      .then(data => { setData(data); setLoading(false); });
  }, [url]);

  return { data, loading };
}`,
		},
		{
			Title: "React: Theme Context",
			Language: "React",
			Difficulty: "hard",
			Content: `const ThemeContext = createContext(undefined);

export function ThemeProvider({ children }) {
  const [theme, setTheme] = useState('dark');

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}`,
		},

		// Java
		{
			Title: "Java: Controller Endpoint",
			Language: "Java",
			Difficulty: "easy",
			Content: `@RestController
@RequestMapping("/api/exercises")
public class ExerciseController {

    private final ExerciseService service;

    @GetMapping("/{id}")
    public ResponseEntity<ExerciseDto> getById(@PathVariable String id) {
        return service.findById(id)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }
}`,
		},
		{
			Title: "Java: Stream Operations",
			Language: "Java",
			Difficulty: "medium",
			Content: `public List<String> getHighPerformanceMembers(List<TropaMember> members) {
    return members.stream()
            .filter(member -> member.getActiveSessionCount() > 15)
            .filter(member -> member.getAverageWpm() >= 70.0)
            .map(TropaMember::getName)
            .sorted()
            .collect(Collectors.toList());
}`,
		},
		{
			Title: "Java: Optional Mapping",
			Language: "Java",
			Difficulty: "hard",
			Content: `@Transactional
public UserProfile updateProfileXp(Long profileId, int xpGained) {
    return repository.findById(profileId)
            .map(profile -> {
                int newXp = profile.getXp() + xpGained;
                profile.setXp(newXp);
                int calculatedLevel = (newXp / 1000) + 1;
                if (calculatedLevel > profile.getLevel()) {
                    profile.setLevel(calculatedLevel);
                }
                return repository.save(profile);
            })
            .orElseThrow(() -> new EntityNotFoundException("Profile not found: " + profileId));
}`,
		},

		// Python
		{
			Title: "Python: List Comprehension",
			Language: "Python",
			Difficulty: "easy",
			Content: `def filter_active_sessions(sessions: list) -> list:
    # Filter sessions with WPM above threshold and return limited view
    active = [s for s in sessions if s.get("wpm", 0) >= 40]
    return [{"id": s["id"], "wpm": s["wpm"]} for s in active[:10]]`,
		},
		{
			Title: "Python: Async Fetching",
			Language: "Python",
			Difficulty: "medium",
			Content: `async def fetch_metric_data(session_id: int) -> dict:
    logging.info(f"Fetching metrics for session {session_id}")
    await asyncio.sleep(0.1) # Simulate NATS communication
    return {
        "session_id": session_id,
        "latency_ms": [140, 185, 205, 98],
        "completed": True
    }`,
		},
		{
			Title: "Python: Decorator Log",
			Language: "Python",
			Difficulty: "hard",
			Content: `def audit_log(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        start_time = time.time()
        logger.info(f"Calling function {func.__name__}")
        try:
            result = func(*args, **kwargs)
            duration = time.time() - start_time
            logger.info(f"Finished {func.__name__} in {duration:.4f}s")
            return result
        except Exception as e:
            logger.error(f"Failed in {func.__name__}: {str(e)}")
            raise
    return wrapper`,
		},
		{
			Title: "Python: Regex Email",
			Language: "Python",
			Difficulty: "easy",
			Content: `import re

def validate_email(email: str) -> bool:
    pattern = r"^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$"
    return bool(re.match(pattern, email))`,
		},
		{
			Title: "Python: Context Manager JSON",
			Language: "Python",
			Difficulty: "medium",
			Content: `def save_config(filepath: str, config: dict) -> None:
    try:
        with open(filepath, "w", encoding="utf-8") as f:
            json.dump(config, f, indent=4)
        logger.info("Configuration saved successfully")
    except IOError as e:
        logger.error(f"Failed to write config file: {e}")`,
		},

		// SQL / JS
		{
			Title: "SQL: JOIN aggregate",
			Language: "SQL",
			Difficulty: "easy",
			Content: `SELECT p.name, COUNT(s.id) AS sessions_completed, AVG(s.wpm) AS avg_speed
FROM profiles p
JOIN sessions s ON p.id = s.profile_id
WHERE s.completed_at >= DATE('now', '-7 days')
GROUP BY p.id
HAVING AVG(s.accuracy) >= 0.95;`,
		},
		{
			Title: "SQL: CTE Ranking",
			Language: "SQL",
			Difficulty: "medium",
			Content: `WITH RankedSessions AS (
    SELECT 
        profile_id,
        wpm,
        accuracy,
        ROW_NUMBER() OVER (PARTITION BY profile_id ORDER BY wpm DESC, accuracy DESC) as rank
    FROM sessions
)
SELECT profile_id, wpm as best_wpm, accuracy as best_accuracy
FROM RankedSessions
WHERE rank = 1;`,
		},
		{
			Title: "SQL: Key latency report",
			Language: "SQL",
			Difficulty: "hard",
			Content: `SELECT 
    km.char,
    SUM(km.attempts) AS total_attempts,
    SUM(km.errors) AS total_errors,
    CAST(SUM(km.errors) AS REAL) / CAST(SUM(km.attempts) AS REAL) * 100 AS error_percentage,
    AVG(km.latency_ms) AS avg_latency_ms
FROM key_metrics km
WHERE km.attempts > 10
GROUP BY km.char
ORDER BY error_percentage DESC, avg_latency_ms DESC
LIMIT 5;`,
		},
		{
			Title: "SQL: Express Web Server",
			Language: "Javascript",
			Difficulty: "easy",
			Content: `const express = require('express');
const app = express();
const port = 3000;

app.get('/api/health', (req, res) => {
  res.json({ status: 'UP', timestamp: new Date() });
});

app.listen(port, () => {
  console.log('Server running');
});`,
		},
		{
			Title: "SQL: Transaction Rollback",
			Language: "SQL",
			Difficulty: "medium",
			Content: `BEGIN TRANSACTION;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;`,
		},
		{
			Title: "SQL: EXISTS Query",
			Language: "SQL",
			Difficulty: "hard",
			Content: `SELECT title, difficulty 
FROM exercises e
WHERE EXISTS (
    SELECT 1 
    FROM sessions s 
    WHERE s.exercise_id = e.id 
      AND s.wpm > 80.0
);`,
		},
	}

	numberPool = []string{
		"192.168.1.105",
		"127.0.0.1:8080",
		"10.0.0.1:27017",
		"8.8.8.8",
		"1.1.1.1",
		"+56 9 9876 5432",
		"2026-06-24",
		"1984-11-23",
		"2018-12-07",
		"30kg barbell",
		"150W switch",
		"41 years old",
		"family of 6",
		"3.14159265",
		"2.71828",
		"100% accuracy",
		"7-string guitar",
		"45.33 WPM",
		"97.8% precision",
		"40 achievements",
		"50 unique lessons",
		"1,000 exercises",
		"Level 28 profile",
		"15,000 XP points",
	}

	symbolPool = []string{
		`if (a && b) { return c ? "yes" : "no"; }`,
		`else if (a || !d) { console.log("error!"); }`,
		`{"status": 200, "ok": true, "data": {"user_id": 101}}`,
		`!@#$%^&*()_+ -={}|[]\:";'<>?,./ ~` + "`" + ` § ±`,
		`<div class="container" id="overlay"> <span *ngIf="key() === '\''"> {{ key }} </span> </div>`,
		`ch <- &User{Name: "Niko", XP: &xpVal}; val := <-ch;`,
		`ptr := *val; addr := &ptr;`,
		`const regex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;`,
		`const [user, setUser] = useState<User | null>(null);`,
		`[ngClass]="{'active': isActive(), 'disabled': !isEnabled()}"`,
		`select * from users where active = 1 and deleted_at is null;`,
		`sudo systemctl restart docker.service && tail -f /var/log/syslog`,
		`git commit -m "feat: setup initial config" && git push origin main`,
		`curl -X POST -H "Content-Type: application/json" -d '{"xp": 100}' https://api.latropa.type`,
		`const { name, level, ...rest } = activeProfile();`,
		`const paths = glob.sync('**/src/**/*.spec.ts');`,
		`func (s *Service) Load(ctx context.Context) (<-chan *Data, error)`,
		`const server = http.createServer((req, res) => { res.end('OK'); });`,
		`{ "presets": ["@babel/preset-env", "@babel/preset-typescript"] }`,
		`const average = arr.reduce((acc, curr) => acc + curr, 0) / arr.length;`,
		`def test_endpoint(client) -> None: assert client.get("/health").status_code == 200`,
		`const config = { host: process.env.DB_HOST || 'localhost', port: 5432 };`,
		`array.filter(x => x !== null).map(x => x.trim().toUpperCase());`,
		`struct KeyMetric { attempts: i32, errors: i32, latency_ms: f32 }`,
		`const result = await Promise.all([p1, p2, p3]);`,
	}
)

type CodeSnippet struct {
	Title      string
	Language   string
	Difficulty string
	Content    string
}

func generateSpanishExercises(count int, rng *rand.Rand) []TempExercise {
	var list []TempExercise
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

		var sentencesNeeded int
		if isEndurance == 1 {
			sentencesNeeded = 8 + rng.Intn(3) // 8 a 10 oraciones
		} else if difficulty == "easy" {
			sentencesNeeded = 2
		} else if difficulty == "medium" {
			sentencesNeeded = 3
		} else {
			sentencesNeeded = 4
		}

		indices := rng.Perm(len(spanishSentences))
		var contentBuilder []string
		var selectedIndices []int
		for k := 0; k < sentencesNeeded && k < len(indices); k++ {
			contentBuilder = append(contentBuilder, spanishSentences[indices[k]])
			selectedIndices = append(selectedIndices, indices[k])
		}

		list = append(list, TempExercise{
			Title:       "Español: Lección",
			Content:     strings.Join(contentBuilder, " "),
			Category:    "spanish",
			Difficulty:  difficulty,
			IsEndurance: isEndurance,
			SourceIDs:   selectedIndices,
		})
	}
	return list
}

func generateEnglishExercises(count int, rng *rand.Rand) []TempExercise {
	var list []TempExercise
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

		var sentencesNeeded int
		if isEndurance == 1 {
			sentencesNeeded = 8 + rng.Intn(3) // 8 to 10 sentences
		} else if difficulty == "easy" {
			sentencesNeeded = 2
		} else if difficulty == "medium" {
			sentencesNeeded = 3
		} else {
			sentencesNeeded = 4
		}

		indices := rng.Perm(len(englishSentences))
		var contentBuilder []string
		var selectedIndices []int
		for k := 0; k < sentencesNeeded && k < len(indices); k++ {
			contentBuilder = append(contentBuilder, englishSentences[indices[k]])
			selectedIndices = append(selectedIndices, indices[k])
		}

		list = append(list, TempExercise{
			Title:       "English: Lesson",
			Content:     strings.Join(contentBuilder, " "),
			Category:    "english",
			Difficulty:  difficulty,
			IsEndurance: isEndurance,
			SourceIDs:   selectedIndices,
		})
	}
	return list
}

func generateCodeExercises(count int, rng *rand.Rand) []TempExercise {
	var list []TempExercise
	for i := 0; i < len(codeSnippets); i++ {
		snippet := codeSnippets[i]
		isEndurance := 0
		difficulty := snippet.Difficulty
		if difficulty == "hard" && rng.Float64() < 0.3 {
			isEndurance = 1
		}

		list = append(list, TempExercise{
			Title:       snippet.Title,
			Content:     snippet.Content,
			Category:    "code",
			Difficulty:  difficulty,
			IsEndurance: isEndurance,
			SourceIDs:   []int{i},
		})
	}
	return list
}

func generateNumbersExercises(count int, rng *rand.Rand) []TempExercise {
	var list []TempExercise
	for i := 0; i < count; i++ {
		var difficulty string
		if i%3 == 0 {
			difficulty = "easy"
		} else if i%3 == 1 {
			difficulty = "medium"
		} else {
			difficulty = "hard"
		}

		var entriesNeeded int
		if difficulty == "easy" {
			entriesNeeded = 3
		} else if difficulty == "medium" {
			entriesNeeded = 5
		} else {
			entriesNeeded = 8
		}

		indices := rng.Perm(len(numberPool))
		var contentBuilder []string
		var selectedIndices []int
		for k := 0; k < entriesNeeded && k < len(indices); k++ {
			contentBuilder = append(contentBuilder, numberPool[indices[k]])
			selectedIndices = append(selectedIndices, indices[k])
		}

		list = append(list, TempExercise{
			Title:       "Números: Lección",
			Content:     strings.Join(contentBuilder, ", "),
			Category:    "numbers",
			Difficulty:  difficulty,
			IsEndurance: 0,
			SourceIDs:   selectedIndices,
		})
	}
	return list
}

func generateSymbolsExercises(count int, rng *rand.Rand) []TempExercise {
	var list []TempExercise
	
	// 1. Easy (25 lecciones)
	for i := 0; i < len(symbolPool); i++ {
		list = append(list, TempExercise{
			Title:       "Símbolos: Lección",
			Content:     symbolPool[i],
			Category:    "symbols",
			Difficulty:  "easy",
			IsEndurance: 0,
			SourceIDs:   []int{i},
		})
	}

	// 2. Medium (25 lecciones)
	for i := 0; i < len(symbolPool); i++ {
		idx1 := i
		idx2 := (i + 5) % len(symbolPool)
		list = append(list, TempExercise{
			Title:       "Símbolos: Lección",
			Content:     fmt.Sprintf("%s\n%s", symbolPool[idx1], symbolPool[idx2]),
			Category:    "symbols",
			Difficulty:  "medium",
			IsEndurance: 0,
			SourceIDs:   []int{idx1, idx2},
		})
	}

	// 3. Hard (25 lecciones)
	for i := 0; i < len(symbolPool); i++ {
		idx1 := i
		idx2 := (i + 7) % len(symbolPool)
		idx3 := (i + 13) % len(symbolPool)
		list = append(list, TempExercise{
			Title:       "Símbolos: Lección",
			Content:     fmt.Sprintf("%s\n%s\n%s", symbolPool[idx1], symbolPool[idx2], symbolPool[idx3]),
			Category:    "symbols",
			Difficulty:  "hard",
			IsEndurance: 0,
			SourceIDs:   []int{idx1, idx2, idx3},
		})
	}
	return list
}
