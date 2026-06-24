# Reglas de Agente para el Workspace

Este archivo contiene reglas de comportamiento específicas del proyecto que el asistente cargará automáticamente al iniciar.

## Lineamientos de Diseño e Interfaz de Usuario
- **Lectura Obligatoria**: Antes de realizar cualquier cambio en el Frontend (Angular: HTML, CSS, TypeScript), debes revisar los lineamientos estéticos definidos en [DESIGN_SYSTEM.md](file:///Users/nparrado/dev/Personal/TypingExcercises/DESIGN_SYSTEM.md).
- **Prohibición de Emoticones (Emojis)**: Está terminantemente prohibido insertar emoticones Unicode de texto (ej: 🔥, 🥇, 🔒, ⚙️, ♥, ⚡, 🎉, 👤) en cualquier plantilla HTML o archivo CSS. Debes usar en su lugar los iconos vectoriales SVG limpios especificados en el catálogo de diseño.
- **Espacio y Composición**: Mantén un diseño limpio, profesional y espacioso. La configuración de usuario (estilos y sonido) debe persistir en base de datos y estar contenida en el modal de ajustes ⚙️ en la cabecera, liberando espacio en el área de lecciones y estadísticas.
- **Velocímetros**: El WPM de las categorías en la pestaña de Estadísticas debe mostrarse mediante arcos SVG dinámicos que representen velocímetros hasta un rango de 150 WPM.
