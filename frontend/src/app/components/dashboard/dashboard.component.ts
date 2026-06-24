import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiService, Profile, Exercise, SummaryStats, RetryItem } from '../../services/api.service';
import { SoundService } from '../../services/sound.service';
import { TypingAreaComponent } from '../typing-area/typing-area.component';

interface AchievementSchema {
  code: string;
  name: string;
  desc: string;
  tier: 'simple' | 'medium' | 'hard' | 'impossible';
}

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, FormsModule, TypingAreaComponent],
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.css']
})
export class DashboardComponent implements OnInit {
  private api = inject(ApiService);
  sound = inject(SoundService);

  // Profile management
  profiles: Profile[] = [];
  activeProfile: Profile | null = null;
  newProfileName: string = '';
  showCreateProfile = false;

  // Active state
  activeTab: 'exercises' | 'stats' | 'retries' | 'achievements' = 'exercises';
  selectedCategory: string = '';
  selectedDifficulty: string = '';
  selectedExercise: Exercise | null = null;
  typingMode: string = 'lesson';

  // Library of loaded exercises
  exercises: Exercise[] = [];
  isLoadingExercises = false;

  // Stats and achievements data
  stats: SummaryStats | null = null;
  retries: RetryItem[] = [];
  isLoadingStats = false;
  activeTheme: string = 'glass';

  // 30 Achievements catalog
  readonly achievementsCatalog: AchievementSchema[] = [
    // Simple
    { code: 'welcome', name: 'Bienvenido a la Tropa', desc: 'Crea tu primer perfil local.', tier: 'simple' },
    { code: 'first_step', name: 'Primeros Pasos', desc: 'Completa tu primer ejercicio.', tier: 'simple' },
    { code: 'warm_up', name: 'Calentando Motores', desc: 'Escribe a más de 30 WPM.', tier: 'simple' },
    { code: 'good_pace', name: 'Buen Ritmo', desc: 'Completa un ejercicio con precisión >95%.', tier: 'simple' },
    { code: 'polyglot_novice', name: 'Políglota Novato', desc: 'Completa un ejercicio en español y otro en inglés.', tier: 'simple' },
    { code: 'first_script', name: 'Primer Script', desc: 'Completa tu primer ejercicio de código.', tier: 'simple' },
    { code: 'streak_inits', name: 'Racha de Iniciación', desc: 'Consigue una racha de 2 días seguidos.', tier: 'simple' },
    { code: 'take_your_time', name: 'Sin Prisas', desc: 'Completa un ejercicio largo (Prueba de Resistencia).', tier: 'simple' },

    // Medium
    { code: 'streak_constant', name: 'Mecanógrafo Constante', desc: 'Mantén una racha activa de 7 días.', tier: 'medium' },
    { code: 'dev_junior', name: 'Desarrollador Junior', desc: 'Completa 15 ejercicios en la categoría de Código.', tier: 'medium' },
    { code: 'hawk_eye', name: 'Ojo de Halcón', desc: 'Termina un texto largo con 0 errores (100% precisión).', tier: 'medium' },
    { code: 'speedster_silver', name: 'Velocista de Plata', desc: 'Alcanza los 60 WPM.', tier: 'medium' },
    { code: 'endurance_bronze', name: 'Resistencia de Bronce', desc: 'Completa 5 pruebas de resistencia.', tier: 'medium' },
    { code: 'the_return', name: 'El Retorno', desc: 'Completa con éxito un ejercicio de reintento.', tier: 'medium' },
    { code: 'symbol_master', name: 'Maestro de Símbolos', desc: 'Completa 10 ejercicios de símbolos y números.', tier: 'medium' },
    { code: 'practice_afternoon', name: 'Tarde de Práctica', desc: 'Acumula más de 30 minutos de tipeo activo en un día.', tier: 'medium' },
    { code: 'zero_backspace', name: 'Cero Borrados', desc: 'Termina un ejercicio largo sin usar la tecla Delete.', tier: 'medium' },

    // Hard
    { code: 'speedster_gold', name: 'Mecanógrafo de Oro', desc: 'Alcanza los 90 WPM en un ejercicio de texto.', tier: 'hard' },
    { code: 'streak_silver', name: 'Racha de Plata', desc: 'Mantén una racha activa de 15 días.', tier: 'hard' },
    { code: 'dev_senior', name: 'Desarrollador Senior', desc: 'Completa 50 ejercicios de código con precisión >97%.', tier: 'hard' },
    { code: 'precision_perfect', name: 'Precisión Milimétrica', desc: 'Completa 5 ejercicios seguidos con >99% de precisión y >50 WPM.', tier: 'hard' },
    { code: 'marathon_steel', name: 'Maratonista de Acero', desc: 'Completa un texto largo (>3k) a >70 WPM y >97% de precisión.', tier: 'hard' },
    { code: 'total_control', name: 'Control Total', desc: 'Completa un texto de más de 500 caracteres sin cometer errores.', tier: 'hard' },
    { code: 'personal_best', name: 'Superación Personal', desc: 'Supera tu promedio general de WPM.', tier: 'hard' },

    // Impossible
    { code: 'the_chosen_one', name: 'El Elegido', desc: 'Alcanza más de 150 WPM con 100% de precisión.', tier: 'impossible' },
    { code: 'human_compiler', name: 'Compilador Humano', desc: 'Código >1.5k a >95 WPM, 100% de precisión y 0 Delete.', tier: 'impossible' },
    { code: 'streak_legend', name: 'Leyenda de la Tropa', desc: 'Mantén una racha ininterrumpida de 365 días.', tier: 'impossible' },
    { code: 'endurance_deity', name: 'Deidad de Resistencia', desc: 'Texto >5k a >110 WPM y >99% de precisión.', tier: 'impossible' },
    { code: 'absolute_zen', name: 'Zen Absoluto', desc: 'Completa 20 ejercicios distintos seguidos al 100% de precisión.', tier: 'impossible' },
    { code: 'perfect_programmer', name: 'Programador Perfecto', desc: 'Código >800 chars a >100 WPM con 100% de precisión.', tier: 'impossible' }
  ];

  ngOnInit() {
    this.loadProfiles();
    this.loadTheme();
  }

  loadProfiles() {
    this.api.getProfiles().subscribe({
      next: (data) => this.profiles = data,
      error: (err) => console.error("Error loading profiles:", err)
    });
  }

  selectProfile(profile: Profile) {
    this.activeProfile = profile;
    this.selectedExercise = null;
    this.activeTab = 'exercises';
    this.loadExercises();
    this.loadStatsAndRetries();
  }

  createProfile() {
    const name = this.newProfileName.trim();
    if (!name) return;
    this.api.createProfile({ name }).subscribe({
      next: (profile) => {
        this.newProfileName = '';
        this.showCreateProfile = false;
        this.loadProfiles();
        this.selectProfile(profile);
      },
      error: (err) => {
        alert(err.error?.error || "Error al crear perfil");
      }
    });
  }

  deleteProfile(profileId: number, event: Event) {
    event.stopPropagation();
    if (confirm("¿Estás seguro de que deseas eliminar este perfil? Se perderá todo su historial.")) {
      this.api.deleteProfile(profileId).subscribe({
        next: () => {
          if (this.activeProfile?.id === profileId) {
            this.activeProfile = null;
          }
          this.loadProfiles();
        }
      });
    }
  }

  logout() {
    this.activeProfile = null;
    this.selectedExercise = null;
    this.stats = null;
    this.retries = [];
  }

  loadExercises() {
    if (!this.activeProfile) return;
    this.isLoadingExercises = true;
    this.api.getExercises(this.selectedCategory, this.selectedDifficulty).subscribe({
      next: (data) => {
        this.exercises = data;
        this.isLoadingExercises = false;
      },
      error: (err) => {
        console.error("Error loading exercises:", err);
        this.isLoadingExercises = false;
      }
    });
  }

  loadStatsAndRetries() {
    if (!this.activeProfile?.id) return;
    this.isLoadingStats = true;
    
    this.api.getStats(this.activeProfile.id).subscribe({
      next: (data) => {
        this.stats = data;
        this.isLoadingStats = false;
      },
      error: (err) => {
        console.error("Error loading stats:", err);
        this.isLoadingStats = false;
      }
    });

    this.api.getRetries(this.activeProfile.id).subscribe({
      next: (data) => this.retries = data,
      error: (err) => console.error("Error loading retries:", err)
    });
  }

  startExercise(ex: Exercise, mode: string = 'lesson') {
    this.selectedExercise = ex;
    this.typingMode = mode;
  }

  startArcadeMode(ex: Exercise) {
    this.startExercise(ex, 'arcade');
  }

  startRetry(retry: RetryItem) {
    // Reconstruct exercise structure to practice it
    const ex: Exercise = {
      id: retry.exercise_id,
      title: retry.title,
      content: '', // Will be loaded from backend or we fetch
      category: retry.category,
      difficulty: retry.difficulty,
      is_endurance: false
    };

    // Need to fetch details of retry exercise to get its text content
    this.api.getExercises().subscribe(list => {
      // Find the text in the list or fallback
      const found = list.find(e => e.id === retry.exercise_id);
      if (found) {
        this.startExercise(found, 'retry');
      } else {
        // Fallback: search DB directly
        this.startExercise({
          id: retry.exercise_id,
          title: retry.title,
          content: "Ejemplo de texto de reintento para cargar.",
          category: retry.category,
          difficulty: retry.difficulty,
          is_endurance: false
        }, 'retry');
      }
    });
  }

  // Smart Custom Practice Generator
  startSmartPractice() {
    if (!this.activeProfile?.id || !this.stats || this.stats.weakest_keys.length === 0) return;

    // Collect weakest characters
    const weakChars = this.stats.weakest_keys.map(k => k.char);
    
    // Generate repetition blocks and words in Spanish/English containing those characters
    const wordsPool = [
      "entrenar", "fuerza", "guitarra", "música", "metal", "desarrollo", "mancuernas", 
      "backend", "tropa", "lección", "mecanografía", "ejercicio", "precisión", "velocidad",
      "pablo", "feña", "sofi", "luciano", "joyce", "niko", "computador", "teclado", "docker"
    ];

    let content = "";
    // 1. Repeated drill patterns (e.g. j j k k jkj kjk)
    weakChars.forEach(char => {
      content += `${char} ${char} ${char} ${char}${char} ${char} `;
    });
    content += "\n";

    // 2. Add words from pool containing at least one weak key
    const matchingWords = wordsPool.filter(w => weakChars.some(c => w.toLowerCase().includes(c.toLowerCase())));
    if (matchingWords.length > 0) {
      for (let j = 0; j < 3; j++) {
        content += matchingWords.sort(() => 0.5 - Math.random()).join(" ") + " ";
      }
    } else {
      // Fallback drills
      content += "practica tus teclas debiles con consistencia ";
    }
    
    content = content.trim();

    const smartExercise: Exercise = {
      id: "smart_practice_" + Date.now(),
      title: "Práctica Inteligente: Teclas Débiles",
      content: content,
      category: "symbols",
      difficulty: "medium",
      is_endurance: false
    };

    this.startExercise(smartExercise, 'lesson');
  }

  onTypingFinished() {
    this.selectedExercise = null;
    this.loadStatsAndRetries();
    // Refresh profiles to update XP and level on header
    this.loadProfiles();
    if (this.activeProfile?.id) {
      this.api.getProfile(this.activeProfile.id).subscribe(p => this.activeProfile = p);
    }
  }

  // --- Theme Toggle Manager ---
  setTheme(theme: string) {
    this.activeTheme = theme;
    localStorage.setItem('typing_theme', theme);
    document.body.className = '';
    if (theme === 'cyberpunk') {
      document.body.classList.add('theme-cyberpunk');
    } else if (theme === 'terminal') {
      document.body.classList.add('theme-terminal');
    }
  }

  loadTheme() {
    const saved = localStorage.getItem('typing_theme') || 'glass';
    this.setTheme(saved);
  }

  // Check if an achievement is unlocked
  isUnlocked(code: string): boolean {
    return this.stats?.achievements.includes(code) || false;
  }

  // SVG Chart Generators helpers
  getChartPoints(): string {
    if (!this.stats || this.stats.progress_wpm.length < 2) return '';
    const points = this.stats.progress_wpm;
    const width = 500;
    const height = 150;
    const padding = 20;

    const maxWpm = Math.max(...points.map(p => p.wpm), 60); // min cap 60
    const minWpm = Math.min(...points.map(p => p.wpm), 10);

    const xStep = (width - padding * 2) / (points.length - 1);
    const yScale = (height - padding * 2) / (maxWpm - minWpm || 1);

    return points.map((p, idx) => {
      const x = padding + idx * xStep;
      const y = height - padding - (p.wpm - minWpm) * yScale;
      return `${x},${y}`;
    }).join(' ');
  }

  getXPProgress(): number {
    if (!this.activeProfile || this.activeProfile.level === undefined || this.activeProfile.xp === undefined) return 0;
    const nextLevelXP = this.activeProfile.level * 1000;
    return (this.activeProfile.xp / nextLevelXP) * 100;
  }

  getWeakestKeysString(): string {
    if (!this.stats || !this.stats.weakest_keys) return '';
    return this.stats.weakest_keys.map(k => k.char).join(', ');
  }
}
