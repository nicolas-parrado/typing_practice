import { Component, Input, Output, EventEmitter, OnInit, OnDestroy, HostListener, signal, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Exercise, ApiService, SaveSessionResponse } from '../../services/api.service';
import { SoundService } from '../../services/sound.service';
import { KeyboardComponent } from '../keyboard/keyboard.component';

interface TelemetryPoint {
  time: number;
  wpm: number;
  errors: number;
}

const INTL_SEQUENCES: { [char: string]: { keys: string[], codes: string[] } } = {
  'á': { keys: ["'", "a"], codes: ["Quote", "KeyA"] },
  'é': { keys: ["'", "e"], codes: ["Quote", "KeyE"] },
  'í': { keys: ["'", "i"], codes: ["Quote", "KeyI"] },
  'ó': { keys: ["'", "o"], codes: ["Quote", "KeyO"] },
  'ú': { keys: ["'", "u"], codes: ["Quote", "KeyU"] },
  'ç': { keys: ["'", "c"], codes: ["Quote", "KeyC"] },
  'Á': { keys: ["'", "A"], codes: ["Quote", "KeyA"] },
  'É': { keys: ["'", "E"], codes: ["Quote", "KeyE"] },
  'Í': { keys: ["'", "I"], codes: ["Quote", "KeyI"] },
  'Ó': { keys: ["'", "O"], codes: ["Quote", "KeyO"] },
  'Ú': { keys: ["'", "U"], codes: ["Quote", "KeyU"] },
  'Ç': { keys: ["'", "C"], codes: ["Quote", "KeyC"] },
  
  'ñ': { keys: ["~", "n"], codes: ["IntlBackslash", "KeyN"] }, // ~ is Shift+IntlBackslash
  'ã': { keys: ["~", "a"], codes: ["IntlBackslash", "KeyA"] },
  'õ': { keys: ["~", "o"], codes: ["IntlBackslash", "KeyO"] },
  'Ñ': { keys: ["~", "N"], codes: ["IntlBackslash", "KeyN"] },
  'Ã': { keys: ["~", "A"], codes: ["IntlBackslash", "KeyA"] },
  'Õ': { keys: ["~", "O"], codes: ["IntlBackslash", "KeyO"] },

  'ü': { keys: ["\"", "u"], codes: ["Quote", "KeyU"] }, // " is Shift+Quote
  'ö': { keys: ["\"", "o"], codes: ["Quote", "KeyO"] },
  'ä': { keys: ["\"", "a"], codes: ["Quote", "KeyA"] },
  'Ü': { keys: ["\"", "U"], codes: ["Quote", "KeyU"] },
  'Ö': { keys: ["\"", "O"], codes: ["Quote", "KeyO"] },
  'Ä': { keys: ["\"", "A"], codes: ["Quote", "KeyA"] },

  'â': { keys: ["^", "a"], codes: ["Digit6", "KeyA"] }, // ^ is Shift+6
  'ê': { keys: ["^", "e"], codes: ["Digit6", "KeyE"] },
  'î': { keys: ["^", "i"], codes: ["Digit6", "KeyI"] },
  'ô': { keys: ["^", "o"], codes: ["Digit6", "KeyO"] },
  'û': { keys: ["^", "u"], codes: ["Digit6", "KeyU"] },
  'Â': { keys: ["^", "A"], codes: ["Digit6", "KeyA"] },
  'Ê': { keys: ["^", "E"], codes: ["Digit6", "KeyE"] },
  'Î': { keys: ["^", "I"], codes: ["Digit6", "KeyI"] },
  'Ô': { keys: ["^", "O"], codes: ["Digit6", "KeyO"] },
  'Û': { keys: ["^", "U"], codes: ["Digit6", "KeyU"] },

  'à': { keys: ["`", "a"], codes: ["IntlBackslash", "KeyA"] },
  'è': { keys: ["`", "e"], codes: ["IntlBackslash", "KeyE"] },
  'ì': { keys: ["`", "i"], codes: ["IntlBackslash", "KeyI"] },
  'ò': { keys: ["`", "o"], codes: ["IntlBackslash", "KeyO"] },
  'ù': { keys: ["`", "u"], codes: ["IntlBackslash", "KeyU"] },
  'À': { keys: ["`", "A"], codes: ["IntlBackslash", "KeyA"] },
  'È': { keys: ["`", "E"], codes: ["IntlBackslash", "KeyE"] },
  'Ì': { keys: ["`", "I"], codes: ["IntlBackslash", "KeyI"] },
  'Ò': { keys: ["`", "O"], codes: ["IntlBackslash", "KeyO"] },
  'Ù': { keys: ["`", "U"], codes: ["IntlBackslash", "KeyU"] },

  "'": { keys: ["'", " "], codes: ["Quote", "Space"] },
  '"': { keys: ["\"", " "], codes: ["Quote", "Space"] },
  '`': { keys: ["`", " "], codes: ["IntlBackslash", "Space"] },
  '~': { keys: ["~", " "], codes: ["IntlBackslash", "Space"] },
  '^': { keys: ["^", " "], codes: ["Digit6", "Space"] }
};

@Component({
  selector: 'app-typing-area',
  standalone: true,
  imports: [CommonModule, KeyboardComponent],
  templateUrl: './typing-area.component.html',
  styleUrls: ['./typing-area.component.css']
})
export class TypingAreaComponent implements OnInit, OnDestroy {
  @Input() exercise!: Exercise;
  @Input() mode: string = 'lesson'; // 'lesson', 'arcade', 'endurance', 'retry'
  @Input() profileId!: number;
  @Output() onFinish = new EventEmitter<void>();

  private api = inject(ApiService);
  private sound = inject(SoundService);

  // States
  isPlaying = false;
  isFinished = false;
  cursorIndex = 0;
  errorsMap: { [index: number]: boolean } = {};
  typedHistory: string[] = [];

  // Ghost / Personal Best Target Tracker
  ghostTargetWpm = 25;
  ghostTargetAccuracy = 90;
  ghostDiffWpm = 0;
  ghostDiffAccuracy = 0;
  hasGhostRecord = false;
  targetGoalWpm = 25;
  minGoalAccuracy = 90;
  coachMessage = '';
  coachWeakKeys: { char: string; errors: number }[] = [];

  // Key tracking
  expectedKey = '';
  expectedFinger = '';
  shiftRequired = false;
  altGrRequired = false;

  // Dead Key states
  deadKeyActive = false;
  deadKeySequence: { keys: string[], codes: string[] } | null = null;
  deadKeyStep = 0;

  // Stats
  wpm = 0;
  accuracy = 100;
  durationSeconds = 0;
  errorsCount = 0;
  backspacesUsed = 0;
  correctCharsCount = 0;
  
  // Arcade mode
  arcadeLives = 3;

  // Timers and Telemetry
  private gameTimer: any = null;
  private telemetryTimer: any = null;
  private startTime: number = 0;
  private keyLatencies: { [key: string]: { start: number, count: number, total: number } } = {};
  
  telemetry: TelemetryPoint[] = [];
  keyMetricsPayload: { [key: string]: { attempts: number, errors: number, latency_ms: number } } = {};

  // Saving state
  isSaving = false;
  saveResponse: SaveSessionResponse | null = null;
  passedGoal = false;

  correctStreak = 0;

  getComboClass(): string {
    if (this.correctStreak >= 40) return 'combo-fire';
    if (this.correctStreak >= 20) return 'combo-purple';
    if (this.correctStreak >= 10) return 'combo-blue';
    return '';
  }

  ngOnInit() {
    this.resetTest();
  }

  ngOnDestroy() {
    this.stopTimers();
  }

  resetTest() {
    this.stopTimers();
    this.isPlaying = false;
    this.isFinished = false;
    this.cursorIndex = 0;
    this.wpm = 0;
    this.accuracy = 100;
    this.durationSeconds = 0;
    this.errorsCount = 0;
    this.backspacesUsed = 0;
    this.correctCharsCount = 0;
    this.errorsMap = {};
    this.typedHistory = [];
    this.arcadeLives = this.mode === 'arcade' ? 3 : 999;
    this.telemetry = [];
    this.keyLatencies = {};
    this.keyMetricsPayload = {};
    this.deadKeyActive = false;
    this.deadKeySequence = null;
    this.deadKeyStep = 0;
    this.isSaving = false;
    this.saveResponse = null;
    this.passedGoal = false;
    this.correctStreak = 0;
    this.coachMessage = '';
    this.coachWeakKeys = [];

    // Load Adaptive Goals & Ghost Target details
    this.targetGoalWpm = this.exercise.target_wpm || 25;
    this.minGoalAccuracy = Math.round((this.exercise.min_accuracy || 0.90) * 100);

    const bestWpm = this.mode === 'arcade' ? (this.exercise.arcade_wpm || 0) : (this.exercise.high_score_wpm || 0);
    const bestAcc = this.mode === 'arcade' ? (this.exercise.arcade_accuracy || 0) : (this.exercise.high_score_accuracy || 0);
    if (bestWpm > 0) {
      this.ghostTargetWpm = bestWpm;
      this.ghostTargetAccuracy = Math.round(bestAcc * 100);
      this.hasGhostRecord = true;
    } else {
      this.ghostTargetWpm = this.targetGoalWpm;
      this.ghostTargetAccuracy = this.minGoalAccuracy;
      this.hasGhostRecord = false;
    }
    this.ghostDiffWpm = 0;
    this.ghostDiffAccuracy = 0;

    this.updateTargetKey();
  }

  startTest() {
    this.isPlaying = true;
    this.startTime = Date.now();
    this.startTimers();
    this.recordKeyStart();
  }

  private startTimers() {
    this.gameTimer = setInterval(() => {
      this.durationSeconds = Math.floor((Date.now() - this.startTime) / 1000);
      this.calculateStats();
    }, 1000);

    // Save telemetry point every 2 seconds
    this.telemetryTimer = setInterval(() => {
      this.telemetry.push({
        time: this.durationSeconds,
        wpm: this.wpm,
        errors: this.errorsCount
      });
    }, 2000);
  }

  private stopTimers() {
    if (this.gameTimer) clearInterval(this.gameTimer);
    if (this.telemetryTimer) clearInterval(this.telemetryTimer);
  }

  private calculateStats() {
    const elapsedMinutes = this.durationSeconds / 60;
    if (elapsedMinutes > 0) {
      // WPM = (correct words / elapsed time). A word is 5 characters.
      this.wpm = Math.round((this.correctCharsCount / 5) / elapsedMinutes);
    } else {
      this.wpm = 0;
    }

    const totalAttempts = this.correctCharsCount + this.errorsCount;
    this.accuracy = totalAttempts > 0 ? Math.round((this.correctCharsCount / totalAttempts) * 100) : 100;

    // Calculate Ghost relative comparison
    this.ghostDiffWpm = this.wpm - this.ghostTargetWpm;
    this.ghostDiffAccuracy = this.accuracy - this.ghostTargetAccuracy;
  }

  private recordKeyStart() {
    const char = this.getCurrentChar();
    if (!this.keyLatencies[char]) {
      this.keyLatencies[char] = { start: Date.now(), count: 0, total: 0 };
    } else {
      this.keyLatencies[char].start = Date.now();
    }
  }

  private recordKeyEnd(char: string, isError: boolean) {
    if (!this.keyMetricsPayload[char]) {
      this.keyMetricsPayload[char] = { attempts: 0, errors: 0, latency_ms: 0 };
    }
    
    this.keyMetricsPayload[char].attempts++;
    if (isError) {
      this.keyMetricsPayload[char].errors++;
    }

    const metric = this.keyLatencies[char];
    if (metric && metric.start > 0) {
      const latency = Date.now() - metric.start;
      metric.total += latency;
      metric.count++;
      this.keyMetricsPayload[char].latency_ms = Math.round(metric.total / metric.count);
    }
  }

  getCurrentChar(): string {
    if (this.cursorIndex >= this.exercise.content.length) return '';
    return this.exercise.content[this.cursorIndex];
  }

  private updateTargetKey() {
    const char = this.getCurrentChar();
    if (!char) return;

    // Check if character requires dead key sequence
    if (INTL_SEQUENCES[char]) {
      this.deadKeySequence = INTL_SEQUENCES[char];
      this.deadKeyActive = true;
      this.expectedKey = this.deadKeySequence.keys[this.deadKeyStep];
    } else {
      this.deadKeyActive = false;
      this.deadKeySequence = null;
      this.deadKeyStep = 0;
      this.expectedKey = char;
    }

    this.mapFingerAndModifiers(this.expectedKey);
  }

  private mapFingerAndModifiers(keyChar: string) {
    this.shiftRequired = false;
    this.altGrRequired = false;

    // Direct key matches or upper case shifts
    if (keyChar === ' ') {
      this.expectedFinger = 'thumb';
      return;
    }

    // Determine Shift requirements
    const isUpper = (keyChar: string) => {
      return keyChar !== keyChar.toLowerCase() && keyChar === keyChar.toUpperCase();
    }

    if (isUpper(keyChar) && keyChar.match(/[A-Z]/)) {
      this.shiftRequired = true;
    }

    // Special symbol shifts
    const shiftSymbols = '!@#$%^&*()_+{}|:"<>?~';
    if (shiftSymbols.includes(keyChar)) {
      this.shiftRequired = true;
    }

    // Map characters to virtual finger IDs
    const leftPinky = 'qaz1!§±`~';
    const leftRing = 'wsx2@';
    const leftMiddle = 'edc3#';
    const leftIndex = 'rfvtgb4$5%';
    const rightIndex = 'yhjnum6^7&';
    const rightMiddle = 'ik,8*';
    const rightRing = 'ol.9(';
    const rightPinky = 'p;:[]{}0)-_=+deleteenter\'"\\|/?';

    const cleanChar = keyChar.toLowerCase();

    if (leftPinky.includes(cleanChar)) this.expectedFinger = 'left-pinky';
    else if (leftRing.includes(cleanChar)) this.expectedFinger = 'left-ring';
    else if (leftMiddle.includes(cleanChar)) this.expectedFinger = 'left-middle';
    else if (leftIndex.includes(cleanChar)) this.expectedFinger = 'left-index';
    else if (rightIndex.includes(cleanChar)) this.expectedFinger = 'right-index';
    else if (rightMiddle.includes(cleanChar)) this.expectedFinger = 'right-middle';
    else if (rightRing.includes(cleanChar)) this.expectedFinger = 'right-ring';
    else if (rightPinky.includes(cleanChar)) this.expectedFinger = 'right-pinky';
    else this.expectedFinger = '';
  }

  @HostListener('window:keydown', ['$event'])
  handleKeyboardEvent(event: KeyboardEvent) {
    if (!this.isPlaying && !this.isFinished && !this.isSaving) {
      if (event.key === 'Enter' || event.key === ' ') {
        event.preventDefault();
        this.startTest();
        return;
      }
    }

    if (!this.isPlaying || this.isFinished || this.isSaving) return;

    // Ignore structural modifier keys alone
    if (['Shift', 'Control', 'Alt', 'Meta', 'CapsLock'].includes(event.key)) {
      return;
    }

    event.preventDefault(); // Stop default browser behaviors (scrolling, page tabs, shortcuts)

    const key = event.key;
    const code = event.code;

    // Handle Backspace
    if (key === 'Backspace') {
      if (this.cursorIndex > 0) {
        this.backspacesUsed++;
        this.cursorIndex--;
        this.typedHistory.pop();
        this.deadKeyStep = 0;
        this.deadKeyActive = false;
        this.correctStreak = 0;
        
        this.sound.playKeySound();
        this.updateTargetKey();
        this.recordKeyStart();
        this.calculateStats();
      }
      return;
    }

    // Get expected char
    const expected = this.getCurrentChar();

    // 1. Double check OS auto-composed input (matches target directly)
    if (key === expected) {
      this.triggerKeystrokeSuccess(expected, event);
      return;
    }

    // 2. Sequence Guide Flow
    if (this.deadKeyActive && this.deadKeySequence) {
      const targetStroke = this.deadKeySequence.keys[this.deadKeyStep];
      const targetCode = this.deadKeySequence.codes[this.deadKeyStep];

      // Validate matching input (with tolerance for Backquote/IntlBackslash swapping on macOS)
      let isMatch = (key === targetStroke || code === targetCode);
      if (!isMatch && (targetCode === 'IntlBackslash' || targetCode === 'Backquote')) {
        if (code === 'IntlBackslash' || code === 'Backquote') {
          isMatch = true;
        }
      }

      if (isMatch) {
        // Correct step in sequence
        this.sound.playKeySound();
        this.createParticle(event);

        if (this.deadKeyStep === 0) {
          // Completed dead key step (e.g. quote pressed)
          this.deadKeyStep = 1;
          this.expectedKey = this.deadKeySequence.keys[1];
          this.mapFingerAndModifiers(this.expectedKey);
        } else {
          // Completed both steps (base letter pressed)
          this.triggerKeystrokeSuccess(expected, event, true);
        }
      } else {
        // Wrong step key pressed
        this.triggerKeystrokeError(expected);
      }
    } else {
      // Normal Key Flow
      if (key === expected || (expected === ' ' && code === 'Space') || (expected === '\n' && code === 'Enter')) {
        this.triggerKeystrokeSuccess(expected, event);
      } else {
        this.triggerKeystrokeError(expected);
      }
    }
  }

  private triggerKeystrokeSuccess(char: string, event: KeyboardEvent, isDeadSequence = false) {
    this.recordKeyEnd(char, false);
    this.sound.playKeySound();
    this.createParticle(event);

    this.typedHistory.push(char);
    this.correctCharsCount++;
    this.cursorIndex++;
    this.correctStreak++;

    // Reset sequence states
    this.deadKeyActive = false;
    this.deadKeyStep = 0;

    this.calculateStats();

    if (this.cursorIndex >= this.exercise.content.length) {
      this.finishTest();
    } else {
      this.updateTargetKey();
      this.recordKeyStart();
    }
  }

  private triggerKeystrokeError(char: string) {
    this.recordKeyEnd(char, true);
    this.sound.playErrorSound();
    
    this.errorsCount++;
    this.errorsMap[this.cursorIndex] = true;
    this.correctStreak = 0;

    if (this.mode === 'arcade') {
      this.arcadeLives--;
      if (this.arcadeLives <= 0) {
        this.finishTest();
      }
    }

    this.calculateStats();
  }

  private finishTest() {
    this.isFinished = true;
    this.isPlaying = false;
    this.stopTimers();
    this.sound.playSuccessSound();
    this.targetGoalWpm = this.exercise.target_wpm || 25;
    this.minGoalAccuracy = Math.round((this.exercise.min_accuracy || 0.90) * 100);
    this.passedGoal = this.wpm >= this.targetGoalWpm && this.accuracy >= this.minGoalAccuracy;
    this.generateCoachDiagnosis();
    this.saveSessionMetrics();
  }

  private generateCoachDiagnosis() {
    // Identify troublesome keys in this session
    const errorsList: { char: string; errors: number }[] = [];
    for (const [ch, stat] of Object.entries(this.keyMetricsPayload)) {
      if (stat.errors > 0) {
        errorsList.push({ char: ch, errors: stat.errors });
      }
    }
    errorsList.sort((a, b) => b.errors - a.errors);
    this.coachWeakKeys = errorsList.slice(0, 3);

    // Synthesize pedagogical coach advice
    if (this.accuracy < this.minGoalAccuracy) {
      this.coachMessage = `Prioriza la precisión antes que la velocidad. Es normal querer escribir rápido, pero la memoria muscular se construye con movimientos limpios y pausados. Reduce el ritmo un 20% en la siguiente ronda.`;
    } else if (this.wpm < this.targetGoalWpm) {
      this.coachMessage = `¡Excelente nivel de precisión (${this.accuracy}%)! Tu técnica es limpia. Ahora busca soltar la tensión en las manos y mantener un flujo continuo entre palabras sin pausas prolongadas.`;
    } else {
      this.coachMessage = `¡Rendimiento sobresaliente! Superaste el objetivo con ${this.wpm} WPM y ${this.accuracy}% de precisión. Tu coordinación motora está en el punto óptimo para la siguiente etapa.`;
    }
  }

  private saveSessionMetrics() {
    this.isSaving = true;

    // Prepare payload
    const payload = {
      profile_id: this.profileId,
      exercise_id: this.exercise.id,
      mode: this.mode,
      wpm: this.wpm,
      accuracy: this.accuracy / 100, // DB expects decimal fraction
      duration_seconds: this.durationSeconds,
      raw_data_json: JSON.stringify(this.telemetry),
      backspaces_used: this.backspacesUsed,
      keys: this.keyMetricsPayload
    };

    this.api.saveSession(payload).subscribe({
      next: (res) => {
        this.saveResponse = res;
        this.isSaving = false;
      },
      error: (err) => {
        console.error("Error saving session statistics:", err);
        this.isSaving = false;
      }
    });
  }

  // Neon Particle Generator
  private createParticle(event: KeyboardEvent) {
    // Generate floating visual particles for tactile satisfaction
    const area = document.querySelector('.typing-view-area');
    if (!area) return;

    const rect = area.getBoundingClientRect();
    const container = document.createElement('div');
    container.className = 'particle-container';
    
    // Spawn near the typing visual text area
    const cursor = document.querySelector('.cursor-highlight');
    let x = rect.left + rect.width / 2;
    let y = rect.top + rect.height / 2;

    if (cursor) {
      const cursorRect = cursor.getBoundingClientRect();
      x = cursorRect.left;
      y = cursorRect.top;
    }

    container.style.left = `${x}px`;
    container.style.top = `${y}px`;

    // Dynamic configuration based on WPM
    let particleCount = 4;
    let baseVelocity = 30;

    if (this.wpm >= 120) {
      particleCount = 16;
      baseVelocity = 70;
    } else if (this.wpm >= 90) {
      particleCount = 12;
      baseVelocity = 55;
    } else if (this.wpm >= 60) {
      particleCount = 8;
      baseVelocity = 45;
    } else if (this.wpm >= 30) {
      particleCount = 6;
      baseVelocity = 35;
    }

    // Dynamic color based on streak (combo)
    let color = 'var(--accent)';
    let isGlowing = false;
    let glowColor = '';

    if (this.correctStreak >= 40) {
      color = '#f97316'; // Naranja fuego
      isGlowing = true;
      glowColor = 'rgba(249, 115, 22, 0.8)';
    } else if (this.correctStreak >= 20) {
      color = '#a855f7'; // Púrpura
      isGlowing = true;
      glowColor = 'rgba(168, 85, 247, 0.8)';
    } else if (this.correctStreak >= 10) {
      color = '#10b981'; // Verde
      isGlowing = true;
      glowColor = 'rgba(16, 185, 129, 0.8)';
    }

    // Emit particle nodes radiating outwards
    for (let i = 0; i < particleCount; i++) {
      const p = document.createElement('div');
      p.className = 'particle';
      
      const angle = Math.random() * Math.PI * 2;
      const velocity = baseVelocity + Math.random() * 50;
      const dx = Math.cos(angle) * velocity;
      const dy = Math.sin(angle) * velocity;

      p.style.setProperty('--dx', `${dx}px`);
      p.style.setProperty('--dy', `${dy}px`);
      
      // Apply color and glow
      p.style.backgroundColor = color;
      if (isGlowing) {
        p.style.boxShadow = `0 0 6px ${glowColor}, 0 0 12px ${glowColor}`;
      }

      // Slightly larger particles at high speeds
      if (this.wpm >= 90) {
        p.style.width = '6px';
        p.style.height = '6px';
      }

      container.appendChild(p);
    }

    document.body.appendChild(container);
    setTimeout(() => container.remove(), 400);
  }

  getCharClass(idx: number): string {
    if (idx < this.cursorIndex) {
      return this.errorsMap[idx] ? 'char-typed-error' : 'char-typed-correct';
    } else if (idx === this.cursorIndex) {
      return 'cursor-highlight';
    } else {
      return 'char-untyped';
    }
  }

  getAchievementName(code: string): string {
    const list: { [key: string]: string } = {
      welcome: "Bienvenido a la Tropa (Perfil Creado)",
      first_step: "Primeros Pasos (Completaste 1 ejercicio)",
      warm_up: "Calentando Motores (Superaste 30 WPM)",
      good_pace: "Buen Ritmo (>95% Precisión)",
      polyglot_novice: "Políglota Novato (Práctica ES & EN)",
      first_script: "Primer Script (Ejercicio Código)",
      streak_inits: "Racha de Iniciación (Racha de 2 días)",
      take_your_time: "Sin Prisas (Prueba Resistencia)",
      streak_constant: "Mecanógrafo Constante (Racha de 7 días)",
      dev_junior: "Desarrollador Junior (15 Lecciones de Código)",
      hawk_eye: "Ojo de Halcón (100% Precisión)",
      speedster_silver: "Velocista de Plata (60 WPM)",
      endurance_bronze: "Resistencia de Bronce (5 Pruebas Resistencia)",
      the_return: "El Retorno (Aprobaste Reintento)",
      symbol_master: "Maestro de Símbolos (10 Lecciones Símbolos)",
      practice_afternoon: "Tarde de Práctica (30 Minutos en el día)",
      zero_backspace: "Cero Borrados (Lección sin Delete)",
      speedster_gold: "Mecanógrafo de Oro (90 WPM)",
      streak_silver: "Racha de Plata (15 días seguidos)",
      dev_senior: "Desarrollador Senior (50 Lecciones Código)",
      precision_perfect: "Precisión Milimétrica (5 ejercicios >99% y >50 WPM)",
      marathon_steel: "Maratonista de Acero (Texto largo >3k y >70 WPM)",
      total_control: "Control Total (>500 chars sin errores)",
      personal_best: "Superación Personal (Batiste tu récord)",
      the_chosen_one: "El Elegido (¡150 WPM con 100% Precisión!)",
      human_compiler: "Compilador Humano (Código largo con 0 Delete, >95 WPM, 100% Precisión)",
      streak_legend: "Leyenda de la Tropa (Racha de 365 días)",
      endurance_deity: "Deidad de Resistencia (Largo >5k, >110 WPM, >99% Precisión)",
      absolute_zen: "Zen Absoluto (20 ejercicios seguidos 100% Precisión)",
      perfect_programmer: "Programador Perfecto (Código >800 chars, >100 WPM, 100% Precisión)",
      
      // Expansion 31-45
      master_classic_spa: "Maestría Clásica: Español (10 Lecciones Español Clásico)",
      master_classic_eng: "Maestría Clásica: Inglés (10 Lecciones Inglés Clásico)",
      master_classic_code: "Maestría Clásica: Código (10 Lecciones Código Clásico)",
      master_classic_num: "Maestría Clásica: Números (10 Lecciones Números Clásico)",
      master_classic_sym: "Maestría Clásica: Símbolos (10 Lecciones Símbolos Clásico)",
      master_arcade_spa: "Maestría Arcade: Español (10 Lecciones Español Arcade)",
      master_arcade_eng: "Maestría Arcade: Inglés (10 Lecciones Inglés Arcade)",
      master_arcade_code: "Maestría Arcade: Código (10 Lecciones Código Arcade)",
      master_arcade_num: "Maestría Arcade: Números (10 Lecciones Números Arcade)",
      master_arcade_sym: "Maestría Arcade: Símbolos (10 Lecciones Símbolos Arcade)",
      elite_code_speed: "Código Limpio Pro (Código >80 WPM, 100% Precisión)",
      elite_spanish_speed: "Furia Española (Español >110 WPM)",
      elite_english_speed: "Ciclón Inglés (Inglés >110 WPM)",
      elite_endurance: "Resistencia de Hierro (3 Lecciones Resistencia >98% Precisión en un día)",
      deity_tropa: "Deidad de la Tropa (Desbloqueaste 40 logros)"
    };
    return list[code] || code;
  }
}
