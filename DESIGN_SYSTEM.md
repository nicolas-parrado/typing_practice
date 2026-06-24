# Lineamientos de Diseño Estético - La Tropa Type

Este documento establece las reglas visuales y de interacción obligatorias para el desarrollo de la interfaz de usuario. Cualquier modificación en el Frontend (Angular) o en los archivos de estilos (CSS) debe ceñirse estrictamente a estas directrices.

---

## 🚫 Regla Fundamental: Cero Emoticonos (Emojis)
Está **estrictamente prohibido** utilizar emoticonos basados en texto o caracteres Unicode (ej. `🔥`, `🥇`, `👤`, `🔒`, `⚙️`, `♥`, `⚡`, `💀`, `🥉`, `🎉`, `↵`, `×`, `✓`, `⭐`).

En su lugar, se deben utilizar **únicamente iconos SVG vectoriales profesionales** con las siguientes pautas:
- Dibujados en línea de trazo limpio (`stroke-linecap="round" stroke-linejoin="round"`).
- Estilo minimalista y moderno (trazo base de `1.5` o `2` píxeles de grosor).
- Colores vinculados a las variables CSS del tema actual (`currentColor` o variables explícitas como `var(--accent)`).

---

## 📂 Biblioteca de Iconos SVG Estandarizados
Usa las siguientes estructuras SVG directamente en los componentes para garantizar uniformidad:

### 1. Candado (Lecciones Bloqueadas)
```html
<svg class="svg-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
  <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
</svg>
```

### 2. Engranaje (Configuraciones)
```html
<svg class="svg-icon" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <circle cx="12" cy="12" r="3"></circle>
  <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
</svg>
```

### 3. Fuego (Rachas / Streaks)
```html
<svg class="svg-icon svg-fire" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <path d="M8.5 14.5A2.5 2.5 0 0 0 11 12c0-1.38-.5-2-1-3-1.072-2.143-.224-4.054 2-6 .5 2.5 2 4.9 4 6.5 2 1.6 3 3.5 3 5.5a7 7 0 1 1-14 0c0-1.153.433-2.294 1-3a2.5 2.5 0 0 0 2.5 2.5z"></path>
</svg>
```

### 4. Corazón (Vidas Arcade)
```html
<svg class="svg-icon svg-heart" width="20" height="20" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
</svg>
```

### 5. Trofeo/Medallas (Logros)
```html
<svg class="svg-icon" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6"></path>
  <path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18"></path>
  <path d="M4 22h16"></path>
  <path d="M10 14.66V17c0 .55-.45 1-1 1H4v2h16v-2h-5c-.55 0-1-.45-1-1v-2.34"></path>
  <path d="M12 2a7 7 0 0 1 7 7c0 2.47-1.19 4.36-3.13 5.34A2 2 0 0 1 14 16.1v.9c0 .55-.45 1-1 1h-2c-.55 0-1-.45-1-1v-.9a2 2 0 0 1-1.87-1.76C6.19 13.36 5 11.47 5 9a7 7 0 0 1 7-7z"></path>
</svg>
```

### 6. Usuario (Perfiles)
```html
<svg class="svg-icon" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
  <circle cx="12" cy="7" r="4"></circle>
</svg>
```

### 7. Play/Iniciar
```html
<svg class="svg-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <polygon points="5 3 19 12 5 21 5 3"></polygon>
</svg>
```

### 8. Cruz (Eliminar / Cerrar / Cancelar)
```html
<svg class="svg-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <line x1="18" y1="6" x2="6" y2="18"></line>
  <line x1="6" y1="6" x2="18" y2="18"></line>
</svg>
```

### 9. Check / Éxito
```html
<svg class="svg-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
  <polyline points="20 6 9 17 4 12"></polyline>
</svg>
```

---

## 🎨 Paleta de Colores por Temas
Cualquier color personalizado debe utilizar las variables de estilo global mapeadas en `styles.css`.
- **Glassmorphism**: Fondo radial indigo/slate, acento violeta (`#6366f1`).
- **Cyberpunk**: Fondo oscuro, acentos en fucsia (`#ff007f`) y cian (`#00f0ff`).
- **Terminal**: Negro puro (`#000000`), textos y acentos en verde fósforo (`#00ff00`).

---

## 📊 Velocímetros (WPM Speedometer)
Los velocímetros de las categorías deben renderizarse usando un arco SVG dinámico (`stroke-dasharray` y `stroke-dashoffset`) con los siguientes umbrales visuales de WPM:
- **0 a 30 WPM (Aprendiz)**: `#ef4444` (Rojo).
- **30 a 70 WPM (Intermedio)**: `#fbbf24` (Amarillo).
- **70 a 110 WPM (Profesional)**: `#10b981` (Verde).
- **110 a 150+ WPM (Leyenda)**: `#00f0ff` (Cian con glow).
