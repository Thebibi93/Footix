# Footix client

Client React/Vite pour une application de pronostics football. Le projet a été migré de Tailwind vers Mantine pour réduire la verbosité des composants et centraliser davantage le système d'interface.

## Stack

- React 19
- Vite 8
- React Router DOM 7
- Mantine 9

## Installation

```bash
npm install
npm run dev
```

## Scripts

- `npm run dev` : lance le serveur de développement Vite
- `npm run build` : construit la version de production
- `npm run preview` : prévisualise le build de production
- `npm run lint` : lance ESLint

## Principes de la migration Mantine

- Suppression de Tailwind et de son plugin Vite
- Ajout d'un `MantineProvider` global
- Centralisation du thème dans `src/theme.js`
- Remplacement des `className` Tailwind par des composants Mantine (`Card`, `Stack`, `Group`, `Button`, `TextInput`, etc.)
- Réduction du CSS global à quelques règles de base

Voir `MIGRATION_MANTINE.md` pour le détail fichier par fichier.
