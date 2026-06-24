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
  selectedCategory: string = 'spanish'; // Default to Spanish
  selectedDifficulty: string = '';
  selectedExercise: Exercise | null = null;
  typingMode: string = 'lesson';

  // Settings modal state
  showSettingsModal = false;

  // Library of loaded exercises
  exercises: Exercise[] = [];
  allExercises: Exercise[] = [];
  isLoadingExercises = false;

  // Stats and achievements data
  stats: SummaryStats | null = null;
  retries: RetryItem[] = [];
  isLoadingStats = false;
  activeTheme: string = 'glass';

  // Category stats for speedometers
  categoryStats: { [key: string]: { wpm: number, accuracy: number, completed: number, total: number } } = {};

  // 45 Achievements catalog
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
    { code: 'perfect_programmer', name: 'Programador Perfecto', desc: 'Código >800 chars a >100 WPM con 100% de precisión.', tier: 'impossible' },

    // Maestría Clásica por Categoría (Medium)
    { code: 'master_classic_spa', name: 'Maestría Clásica: Español', desc: 'Completa 10 lecciones de español en modo clásico.', tier: 'medium' },
    { code: 'master_classic_eng', name: 'Maestría Clásica: Inglés', desc: 'Completa 10 lecciones de inglés en modo clásico.', tier: 'medium' },
    { code: 'master_classic_code', name: 'Maestría Clásica: Código', desc: 'Completa 10 lecciones de código en modo clásico.', tier: 'medium' },
    { code: 'master_classic_num', name: 'Maestría Clásica: Números', desc: 'Completa 10 lecciones de números en modo clásico.', tier: 'medium' },
    { code: 'master_classic_sym', name: 'Maestría Clásica: Símbolos', desc: 'Completa 10 lecciones de símbolos en modo clásico.', tier: 'medium' },

    // Maestría Arcade por Categoría (Hard)
    { code: 'master_arcade_spa', name: 'Maestría Arcade: Español', desc: 'Completa 10 lecciones de español en modo arcade.', tier: 'hard' },
    { code: 'master_arcade_eng', name: 'Maestría Arcade: Inglés', desc: 'Completa 10 lecciones de inglés en modo arcade.', tier: 'hard' },
    { code: 'master_arcade_code', name: 'Maestría Arcade: Código', desc: 'Completa 10 lecciones de código en modo arcade.', tier: 'hard' },
    { code: 'master_arcade_num', name: 'Maestría Arcade: Números', desc: 'Completa 10 lecciones de números en modo arcade.', tier: 'hard' },
    { code: 'master_arcade_sym', name: 'Maestría Arcade: Símbolos', desc: 'Completa 10 lecciones de símbolos en modo arcade.', tier: 'hard' },

    // Élite
    { code: 'elite_code_speed', name: 'Código Limpio Pro', desc: 'Código a más de 80 WPM con 100% de precisión.', tier: 'hard' },
    { code: 'elite_spanish_speed', name: 'Furia Española', desc: 'Supera los 110 WPM en un ejercicio de español.', tier: 'hard' },
    { code: 'elite_english_speed', name: 'Ciclón Inglés', desc: 'Supera los 110 WPM en un ejercicio de inglés.', tier: 'hard' },
    { code: 'elite_endurance', name: 'Resistencia de Hierro', desc: 'Completa 3 lecciones de resistencia con >98% de precisión en un solo día.', tier: 'hard' },
    { code: 'deity_tropa', name: 'Deidad de la Tropa', desc: 'Desbloquea al menos 40 logros.', tier: 'impossible' }
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
    
    // Apply profile settings
    this.activeTheme = profile.theme || 'glass';
    this.setTheme(this.activeTheme);
    this.sound.soundEnabled.set(profile.sound_enabled !== undefined ? profile.sound_enabled : true);
    this.sound.switchType.set((profile.switch_type as any) || 'blue');

    this.loadExercises();
    this.loadStatsAndRetries();
    this.loadAllExercisesForStats();
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
    this.showSettingsModal = false;
  }

  loadExercises() {
    if (!this.activeProfile) return;
    this.isLoadingExercises = true;
    this.api.getExercises(this.selectedCategory, this.selectedDifficulty, undefined, this.activeProfile.id).subscribe({
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

  loadAllExercisesForStats() {
    if (!this.activeProfile?.id) return;
    this.api.getExercises(undefined, undefined, undefined, this.activeProfile.id).subscribe({
      next: (data) => {
        this.allExercises = data;
        this.calculateCategoryStats();
      },
      error: (err) => console.error("Error loading all exercises for stats:", err)
    });
  }

  calculateCategoryStats() {
    const categories = ['spanish', 'english', 'code', 'numbers', 'symbols'];
    this.categoryStats = {};

    categories.forEach(cat => {
      const catExs = this.allExercises.filter(e => e.category === cat);
      const playedExs = catExs.filter(e => (e.high_score_wpm && e.high_score_wpm > 0) || (e.arcade_wpm && e.arcade_wpm > 0));

      const totalWpm = playedExs.reduce((acc, curr) => {
        const best = Math.max(curr.high_score_wpm || 0, curr.arcade_wpm || 0);
        return acc + best;
      }, 0);

      const totalAcc = playedExs.reduce((acc, curr) => {
        const best = Math.max(curr.high_score_accuracy || 0, curr.arcade_accuracy || 0);
        return acc + best;
      }, 0);

      const completed = catExs.filter(e => {
        const hasPassedClassic = (e.high_score_wpm || 0) >= 25 && (e.high_score_accuracy || 0) >= 0.90;
        const hasPassedArcade = (e.arcade_wpm || 0) >= 25 && (e.arcade_accuracy || 0) >= 0.90;
        return hasPassedClassic || hasPassedArcade;
      }).length;

      this.categoryStats[cat] = {
        wpm: playedExs.length > 0 ? Math.round(totalWpm / playedExs.length) : 0,
        accuracy: playedExs.length > 0 ? (totalAcc / playedExs.length) : 0,
        completed: completed,
        total: catExs.length
      };
    });
  }

  getSpeedometerOffset(wpm: number): number {
    const maxWpm = 150;
    const clampedWpm = Math.min(Math.max(wpm, 0), maxWpm);
    const arcLength = 188.5; // Circular arc mapping
    return arcLength - (arcLength * clampedWpm) / maxWpm;
  }

  getSpeedometerColor(wpm: number): string {
    if (wpm < 30) return '#ef4444';      // Red
    if (wpm < 70) return '#fbbf24';      // Yellow
    if (wpm < 110) return '#10b981';     // Green
    return '#00f0ff';                    // Cyan/Glow
  }

  getSpeedometerTier(wpm: number): string {
    if (wpm < 30) return 'Aprendiz';
    if (wpm < 70) return 'Intermedio';
    if (wpm < 110) return 'Profesional';
    return 'Leyenda';
  }

  selectCategory(category: string) {
    this.selectedCategory = category;
    this.loadExercises();
  }

  getCategoryName(cat: string): string {
    const names: { [key: string]: string } = {
      'spanish': 'Español',
      'english': 'Inglés',
      'code': 'Código Real',
      'numbers': 'Números',
      'symbols': 'Símbolos'
    };
    return names[cat] || cat;
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

  saveSettings() {
    if (!this.activeProfile?.id) return;
    const settings = {
      theme: this.activeTheme,
      sound_enabled: this.sound.soundEnabled(),
      switch_type: this.sound.switchType()
    };
    this.api.updateProfileSettings(this.activeProfile.id, settings).subscribe({
      next: () => {
        if (this.activeProfile) {
          this.activeProfile.theme = settings.theme;
          this.activeProfile.sound_enabled = settings.sound_enabled;
          this.activeProfile.switch_type = settings.switch_type;
        }
      },
      error: (err) => console.error("Error updating settings:", err)
    });
  }

  toggleSound() {
    this.sound.soundEnabled.set(!this.sound.soundEnabled());
    this.saveSettings();
  }

  changeSwitchType(type: 'blue' | 'brown' | 'typewriter') {
    this.sound.switchType.set(type);
    this.saveSettings();
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
      content: '',
      category: retry.category,
      difficulty: retry.difficulty,
      is_endurance: false
    };

    this.api.getExercises().subscribe(list => {
      const found = list.find(e => e.id === retry.exercise_id);
      if (found) {
        this.startExercise(found, 'retry');
      } else {
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

  startSmartPractice() {
    if (!this.activeProfile?.id || !this.stats || this.stats.weakest_keys.length === 0) return;

    const weakChars = this.stats.weakest_keys.map(k => k.char);
    const wordsPool = [
      "entrenar", "fuerza", "guitarra", "música", "metal", "desarrollo", "mancuernas",
      "backend", "tropa", "lección", "mecanografía", "ejercicio", "precisión", "velocidad",
      "pablo", "feña", "sofi", "luciano", "joyce", "niko", "computador", "teclado", "docker"
    ];

    let content = "";
    weakChars.forEach(char => {
      content += `${char} ${char} ${char} ${char}${char} ${char} `;
    });
    content += "\n";

    const matchingWords = wordsPool.filter(w => weakChars.some(c => w.toLowerCase().includes(c.toLowerCase())));
    if (matchingWords.length > 0) {
      for (let j = 0; j < 3; j++) {
        content += matchingWords.sort(() => 0.5 - Math.random()).join(" ") + " ";
      }
    } else {
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
    this.loadAllExercisesForStats();
    this.loadProfiles();
    if (this.activeProfile?.id) {
      this.api.getProfile(this.activeProfile.id).subscribe(p => this.activeProfile = p);
    }
    this.loadExercises();
  }

  setTheme(theme: string) {
    this.activeTheme = theme;
    localStorage.setItem('typing_theme', theme);
    document.body.className = '';
    if (theme === 'cyberpunk') {
      document.body.classList.add('theme-cyberpunk');
    } else if (theme === 'terminal') {
      document.body.classList.add('theme-terminal');
    }
    this.saveSettings();
  }

  loadTheme() {
    const saved = localStorage.getItem('typing_theme') || 'glass';
    this.setTheme(saved);
  }

  isUnlocked(code: string): boolean {
    return this.stats?.achievements.includes(code) || false;
  }

  getChartPoints(): string {
    if (!this.stats || this.stats.progress_wpm.length < 2) return '';
    const points = this.stats.progress_wpm;
    const width = 500;
    const height = 150;
    const padding = 20;

    const maxWpm = Math.max(...points.map(p => p.wpm), 60);
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
