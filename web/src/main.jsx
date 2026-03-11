// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import React from 'react';
import ReactDOM from 'react-dom/client';
import { ChakraProvider, extendTheme } from '@chakra-ui/react';
import App from './App';

const theme = extendTheme({
  config: {
    initialColorMode: 'dark',
    useSystemColorMode: false,
  },
  styles: {
    global: {
      body: {
        bg: '#0d1117',
        color: 'white',
      },
    },
  },
  colors: {
    // Kartoza brand colors
    brand: {
      50: '#fdf3e3',
      100: '#fae1b8',
      200: '#f7cd8a',
      300: '#f4b95c',
      400: '#D4922A', // Kartoza orange
      500: '#D4922A', // Kartoza orange (primary)
      600: '#b87e24',
      700: '#9c6a1e',
      800: '#805618',
      900: '#644312',
    },
    kartoza: {
      orange: {
        400: '#D4922A',
        500: '#D4922A',
      },
      blue: {
        300: '#6bb3f0',
        400: '#4a9de8',
        500: '#2a87d8',
      },
    },
    // Semantic background colors
    bg: {
      primary: '#0d1117',
      secondary: '#161b22',
      tertiary: '#21262d',
    },
    accent: {
      red: '#ff4466',
      green: '#4CAF50',
    },
  },
  semanticTokens: {
    colors: {
      'bg.primary': '#0d1117',
      'bg.secondary': '#161b22',
      'bg.tertiary': '#21262d',
    },
  },
  fonts: {
    heading: '"Inter", system-ui, sans-serif',
    body: '"Inter", system-ui, sans-serif',
    mono: '"JetBrains Mono", "Fira Code", monospace',
  },
  components: {
    Button: {
      defaultProps: {
        colorScheme: 'brand',
      },
    },
    Link: {
      baseStyle: {
        _hover: {
          textDecoration: 'none',
        },
      },
    },
  },
});

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ChakraProvider theme={theme}>
      <App />
    </ChakraProvider>
  </React.StrictMode>
);
