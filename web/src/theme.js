import { createTheme } from '@mui/material';

const getTheme = (mode) => createTheme({
  palette: {
    mode,
    primary: {
      main: mode === 'dark' ? '#90caf9' : '#1976d2',
      light: mode === 'dark' ? '#e3f2fd' : '#42a5f5',
      dark: mode === 'dark' ? '#42a5f5' : '#1565c0',
    },
    secondary: {
      main: mode === 'dark' ? '#ce93d8' : '#9c27b0',
      light: mode === 'dark' ? '#f3e5f5' : '#ba68c8',
      dark: mode === 'dark' ? '#ab47bc' : '#7b1fa2',
    },
    background: {
      default: mode === 'dark' ? '#121212' : '#f0f2f5',
      paper: mode === 'dark' ? '#1e1e1e' : '#ffffff',
    },
  },
  components: {
    MuiCard: {
      styleOverrides: {
        root: {
          boxShadow: mode === 'dark' 
            ? '0 2px 4px -1px rgba(0,0,0,0.4), 0 4px 5px 0 rgba(0,0,0,0.35), 0 1px 10px 0 rgba(0,0,0,0.32)'
            : '0 2px 4px -1px rgba(0,0,0,0.2), 0 4px 5px 0 rgba(0,0,0,0.14), 0 1px 10px 0 rgba(0,0,0,0.12)',
          transition: 'transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out',
          '&:hover': {
            transform: 'translateY(-2px)',
            boxShadow: mode === 'dark'
              ? '0 4px 8px -2px rgba(0,0,0,0.5), 0 6px 10px 0 rgba(0,0,0,0.45), 0 2px 15px 0 rgba(0,0,0,0.42)'
              : '0 4px 8px -2px rgba(0,0,0,0.3), 0 6px 10px 0 rgba(0,0,0,0.24), 0 2px 15px 0 rgba(0,0,0,0.22)',
          },
        },
      },
    },
    MuiDrawer: {
      styleOverrides: {
        paper: {
          width: 200,
        },
      },
    },
  },
});

export default getTheme; 