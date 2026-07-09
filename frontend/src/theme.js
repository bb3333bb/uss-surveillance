import { createTheme } from '@mui/material/styles';

const theme = createTheme({
  palette: {
    mode: 'dark',
    background: {
      default: '#0a192f', // Deep Navy
      paper: '#172a45',   // Steel Blue
    },
    primary: {
      main: '#00b0ff',    // Ocean Blue
    },
    secondary: {
      main: '#00e5ff',    // Neon Cyan
    },
    text: {
      primary: '#e6f1ff',
      secondary: '#8892b0',
    },
    warning: {
      main: '#ff9100',    // Bright Orange
    },
    error: {
      main: '#ff3d00',      // Neon Red
    },
  },
  typography: {
    fontFamily: 'Roboto, system-ui, sans-serif',
    h1: { fontWeight: 700 },
    h2: { fontWeight: 600 },
    h3: { fontWeight: 600 },
    h4: { fontWeight: 600 },
    h5: { fontWeight: 500 },
    h6: { fontWeight: 500 },
    body1: { fontSize: '0.925rem' },
    body2: { fontSize: '0.825rem' },
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 600,
          borderRadius: 4,
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
        },
      },
    },
  },
});

export default theme;
