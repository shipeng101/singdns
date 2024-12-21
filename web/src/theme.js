import { createTheme } from '@mui/material/styles';

const getTheme = (mode) => {
  return createTheme({
    palette: {
      mode,
      ...(mode === 'light'
        ? {
            primary: {
              main: '#3949ab',
              light: '#6f74dd',
              dark: '#00227b',
            },
            secondary: {
              main: '#5e35b1',
              light: '#9162e4',
              dark: '#280680',
            },
            background: {
              default: '#f5f7fa',
              paper: '#ffffff',
            },
            text: {
              primary: 'rgba(0, 0, 0, 0.87)',
              secondary: 'rgba(0, 0, 0, 0.6)',
            },
          }
        : {
            primary: {
              main: '#5c6bc0',
              light: '#8e99f3',
              dark: '#26418f',
            },
            secondary: {
              main: '#7e57c2',
              light: '#b085f5',
              dark: '#4d2c91',
            },
            background: {
              default: '#0a1929',
              paper: '#0f2744',
            },
            text: {
              primary: 'rgba(255, 255, 255, 0.87)',
              secondary: 'rgba(255, 255, 255, 0.6)',
            },
          }),
    },
    shape: {
      borderRadius: 8,
    },
    typography: {
      fontFamily: '"Inter", "Roboto", "Helvetica", "Arial", sans-serif',
      h1: {
        fontWeight: 600,
      },
      h2: {
        fontWeight: 600,
      },
      h3: {
        fontWeight: 600,
      },
      h4: {
        fontWeight: 600,
      },
      h5: {
        fontWeight: 600,
      },
      h6: {
        fontWeight: 600,
      },
      subtitle1: {
        fontWeight: 500,
      },
      subtitle2: {
        fontWeight: 500,
      },
      body1: {
        fontSize: '1rem',
        lineHeight: 1.5,
      },
      body2: {
        fontSize: '0.875rem',
        lineHeight: 1.43,
      },
      button: {
        textTransform: 'none',
        fontWeight: 500,
      },
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          body: {
            scrollbarWidth: 'thin',
            scrollbarColor: mode === 'dark' 
              ? 'rgba(255, 255, 255, 0.2) transparent'
              : 'rgba(0, 0, 0, 0.2) transparent',
            '&::-webkit-scrollbar': {
              width: '8px',
              height: '8px',
            },
            '&::-webkit-scrollbar-track': {
              background: 'transparent',
            },
            '&::-webkit-scrollbar-thumb': {
              backgroundColor: mode === 'dark'
                ? 'rgba(255, 255, 255, 0.2)'
                : 'rgba(0, 0, 0, 0.2)',
              borderRadius: '4px',
              '&:hover': {
                backgroundColor: mode === 'dark'
                  ? 'rgba(255, 255, 255, 0.3)'
                  : 'rgba(0, 0, 0, 0.3)',
              },
            },
          },
        },
      },
      MuiCard: {
        styleOverrides: {
          root: {
            backgroundImage: 'none',
          },
        },
      },
      MuiButton: {
        styleOverrides: {
          root: {
            textTransform: 'none',
            borderRadius: '8px',
            fontWeight: 500,
          },
          contained: {
            boxShadow: 'none',
            '&:hover': {
              boxShadow: mode === 'dark'
                ? '0 2px 8px 0 rgba(0,0,0,0.3)'
                : '0 2px 8px 0 rgba(0,0,0,0.2)',
            },
          },
        },
      },
      MuiTextField: {
        defaultProps: {
          variant: 'outlined',
        },
        styleOverrides: {
          root: {
            '& .MuiOutlinedInput-root': {
              borderRadius: '8px',
            },
          },
        },
      },
      MuiDialog: {
        styleOverrides: {
          paper: {
            borderRadius: '12px',
          },
        },
      },
      MuiDivider: {
        styleOverrides: {
          root: {
            opacity: mode === 'dark' ? 0.1 : 0.2,
          },
        },
      },
      MuiListItem: {
        styleOverrides: {
          root: {
            borderRadius: '8px',
          },
        },
      },
      MuiChip: {
        styleOverrides: {
          root: {
            borderRadius: '6px',
          },
        },
      },
      MuiAlert: {
        styleOverrides: {
          root: {
            borderRadius: '8px',
          },
        },
      },
      MuiLinearProgress: {
        styleOverrides: {
          root: {
            borderRadius: '4px',
            overflow: 'hidden',
          },
        },
      },
    },
  });
};

export default getTheme; 