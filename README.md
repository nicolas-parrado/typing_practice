# La Tropa Type - Typing Exercises App

Una aplicación web interactiva, moderna y gamificada diseñada para la práctica diaria de mecanografía, adaptada al diseño físico y combinaciones de teclado **US International** (Mac ISO). 

Desarrollada con un backend robusto en **Golang + SQLite** y un frontend reactivo en **Angular 19** con Signals, completamente containerizada con Docker y servida a través de un proxy inverso con **Nginx**.

---

## 🚀 Características Principales

### 1. Guía Secuencial de Teclas Muertas (US International)
Visualizador interactivo que guía paso a paso al escribir caracteres acentuados o especiales (como `á`, `é`, `ñ`, `ü`, `^`):
- **Guía visual dinámica**: Primero se resalta la tecla muerta correspondiente (ej. `'`, `~`) y el dedo correcto a utilizar. Luego, en el segundo paso, se resalta la letra base.
- **Detección inteligente**: Compatible con la auto-composición nativa del sistema operativo para usuarios experimentados.

### 2. Base de Datos de 1,000+ Ejercicios Reales
El backend en Go utiliza la directiva `//go:embed` para semillar automáticamente la base de datos SQLite en su primer inicio:
- **Código Real**: Fragmentos de código funcionales en **Golang, Angular (TypeScript/HTML), Java, Python y SQL**.
- **Práctica Multilingüe**: Textos e historias coherentes del mundo real en **Español** e **Inglés**.
- **Pruebas de Resistencia**: Ejercicios de larga duración (de 2,000 a 5,000 caracteres) que evalúan la fatiga del mecanógrafo calculando velocidades y errores en tiempo real.

### 3. Gamificación y Sistema de 30 Medallas
- **Rachas Activas**: Control de constancia diaria persistente en base de datos.
- **Logros en 4 Niveles**: Clasificación de 30 logros en niveles *Simple*, *Medio*, *Difícil* y *Casi Imposible* (por ejemplo, completar códigos largos sin usar la tecla de borrado a velocidades extremas).
- **Modo Arcade (Supervivencia)**: Entrena con un límite estricto de 3 vidas, donde cometer errores descuenta corazones.

### 4. Audio Sintetizado (Web Audio API)
Generación de audio dinámica directamente desde el navegador (sin necesidad de descargas externas):
- Clics de interruptores mecánicos Cherry MX Blue (ruidoso) y Brown (táctil).
- Sonidos retro de máquina de escribir mecánica.
- Zumbador de error de baja frecuencia y melodías arpegiadas de éxito al completar lecciones o subir de nivel.

### 5. Multiperfil Familiar Local
Permite cambiar instantáneamente entre perfiles locales sin contraseñas, ideal para hogares con múltiples usuarios (ej. "La Tropa").

### 6. Selector de 3 Temas Premium (Vanilla CSS)
- **Sleek Dark Glassmorphism**: Fondo oscuro degradado con efectos translúcidos de desenfoque.
- **Cyberpunk Neon**: Luces neón cian y violeta que vibran con efectos de brillos.
- **Retro Terminal/Matrix**: Apariencia clásica de terminal informática en fósforo verde.

---

## 🛠️ Arquitectura Técnica

```
                    ┌─────────────────────────┐
                    │     Cliente Angular     │
                    │        (Port 80)        │
                    └────────────┬────────────┘
                                 │ HTTP requests
                                 ▼
                    ┌─────────────────────────┐
                    │   Nginx Reverse Proxy   │
                    │       (Port 8082)       │
                    └────────────┬────────────┘
                                 │
                   ┌─────────────┴─────────────┐
                   │ /api/*                    │ / (Static Files)
                   ▼                           ▼
      ┌─────────────────────────┐ ┌─────────────────────────┐
      │   Go REST API Service   │ │   Angular Static Build  │
      │       (Port 8081)       │ │   (/usr/share/nginx)    │
      └────────────┬────────────┘ └─────────────────────────┘
                   │
                   ▼
      ┌─────────────────────────┐
      │     SQLite Database     │
      │       (typing.db)       │
      └─────────────────────────┘
```

---

## 🐳 Requisitos y Construcción (Docker)

La aplicación está diseñada para ser completamente modular. Ambos contenedores aprovechan las capas de caché de Docker para evitar descargas o instalaciones redundantes de paquetes en compilaciones futuras.

### Levantar la Aplicación
Ejecuta el siguiente comando en la raíz del proyecto para construir y levantar los contenedores:

```bash
docker compose up --build
```

Una vez que el proceso de inicio se complete, abre tu navegador e ingresa a:
👉 **`http://localhost:8082`**

### Puertos en Uso:
- **`8082`**: Frontend (Nginx Proxy). Acceso principal al sitio web.
- **`8081`**: Backend en Go (REST API).

---

## 💻 Desarrollo Local (Sin Docker)

### Requisitos:
- Node.js >= v22
- Go >= 1.22
- SQLite

### Levantar Backend:
```bash
cd backend
go run main.go
```
*El backend escuchará por defecto en el puerto `8081` y creará la base de datos en `./backend/data/typing.db`.*

### Levantar Frontend:
```bash
cd frontend
npm install
npm run start
```
*El frontend de desarrollo se levantará en `http://localhost:4200`.*
