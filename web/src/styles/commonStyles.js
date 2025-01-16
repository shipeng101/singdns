import { alpha } from '@mui/material/styles';

export const getCommonStyles = (theme) => ({
  headerCard: {
    background: theme.palette.mode === 'dark'
      ? 'linear-gradient(to right, #4338ca, #5b21b6)'
      : 'linear-gradient(to right, #4f46e5, #7c3aed)',
    color: '#fff',
    borderRadius: '12px',
    marginBottom: '16px',
    boxShadow: theme.palette.mode === 'dark'
      ? '0 4px 20px 0 rgba(0,0,0,0.25)'
      : '0 4px 20px 0 rgba(0,0,0,0.15)',
    transition: 'all 0.3s ease-in-out',
    '&:hover': {
      transform: 'translateY(-2px)',
      boxShadow: theme.palette.mode === 'dark'
        ? '0 6px 25px 0 rgba(0,0,0,0.3)'
        : '0 6px 25px 0 rgba(0,0,0,0.2)',
    }
  },
  card: {
    borderRadius: '12px',
    background: theme.palette.mode === 'dark'
      ? alpha(theme.palette.background.paper, 0.8)
      : theme.palette.background.paper,
    backdropFilter: 'blur(10px)',
    boxShadow: theme.palette.mode === 'dark'
      ? '0 4px 20px 0 rgba(0,0,0,0.25)'
      : '0 4px 20px 0 rgba(0,0,0,0.15)',
    transition: 'all 0.3s ease-in-out',
    '&:hover': {
      transform: 'translateY(-2px)',
      boxShadow: theme.palette.mode === 'dark'
        ? '0 6px 25px 0 rgba(0,0,0,0.3)'
        : '0 6px 25px 0 rgba(0,0,0,0.2)',
    }
  },
  actionButton: {
    borderRadius: '8px',
    textTransform: 'none',
    fontWeight: 500,
    boxShadow: 'none',
    background: theme.palette.mode === 'dark'
      ? 'linear-gradient(45deg, #1a237e 30%, #311b92 90%)'
      : 'linear-gradient(45deg, #3949ab 30%, #5e35b1 90%)',
    '&:hover': {
      boxShadow: theme.palette.mode === 'dark'
        ? '0 4px 12px 0 rgba(0,0,0,0.3)'
        : '0 4px 12px 0 rgba(0,0,0,0.2)',
    }
  },
  iconButton: {
    color: 'rgba(255,255,255,0.8)',
    '&:hover': {
      backgroundColor: 'rgba(255,255,255,0.1)',
      color: '#fff',
    }
  },
  dialog: {
    borderRadius: '12px',
    background: theme.palette.mode === 'dark'
      ? alpha(theme.palette.background.paper, 0.9)
      : theme.palette.background.paper,
    backdropFilter: 'blur(10px)',
  },
  chip: {
    primary: {
      borderRadius: '6px',
      fontWeight: 500,
      boxShadow: theme.palette.mode === 'dark'
        ? '0 2px 8px 0 rgba(0,0,0,0.2)'
        : '0 2px 8px 0 rgba(0,0,0,0.1)',
    },
    success: {
      borderRadius: '6px',
      fontWeight: 500,
      boxShadow: theme.palette.mode === 'dark'
        ? '0 2px 8px 0 rgba(0,0,0,0.2)'
        : '0 2px 8px 0 rgba(0,0,0,0.1)',
    },
    error: {
      borderRadius: '6px',
      fontWeight: 500,
      boxShadow: theme.palette.mode === 'dark'
        ? '0 2px 8px 0 rgba(0,0,0,0.2)'
        : '0 2px 8px 0 rgba(0,0,0,0.1)',
    },
    warning: {
      borderRadius: '6px',
      fontWeight: 500,
      boxShadow: theme.palette.mode === 'dark'
        ? '0 2px 8px 0 rgba(0,0,0,0.2)'
        : '0 2px 8px 0 rgba(0,0,0,0.1)',
    }
  },
  listItem: {
    borderRadius: '8px',
    transition: 'all 0.2s ease-in-out',
    '&:hover': {
      backgroundColor: theme.palette.mode === 'dark'
        ? alpha(theme.palette.action.hover, 0.1)
        : theme.palette.action.hover,
      transform: 'translateX(4px)',
    }
  },
  textField: {
    '& .MuiOutlinedInput-root': {
      borderRadius: '8px',
      transition: 'all 0.2s ease-in-out',
      '&:hover': {
        backgroundColor: theme.palette.mode === 'dark'
          ? alpha(theme.palette.action.hover, 0.1)
          : theme.palette.action.hover,
      },
      '&.Mui-focused': {
        boxShadow: theme.palette.mode === 'dark'
          ? '0 0 0 2px rgba(255,255,255,0.2)'
          : '0 0 0 2px rgba(0,0,0,0.1)',
      }
    }
  },
  select: {
    '& .MuiOutlinedInput-root': {
      borderRadius: '8px',
      transition: 'all 0.2s ease-in-out',
      '&:hover': {
        backgroundColor: theme.palette.mode === 'dark'
          ? alpha(theme.palette.action.hover, 0.1)
          : theme.palette.action.hover,
      },
      '&.Mui-focused': {
        boxShadow: theme.palette.mode === 'dark'
          ? '0 0 0 2px rgba(255,255,255,0.2)'
          : '0 0 0 2px rgba(0,0,0,0.1)',
      }
    }
  },
  switch: {
    '& .MuiSwitch-switchBase.Mui-checked': {
      color: theme.palette.mode === 'dark'
        ? '#1a237e'
        : '#3949ab',
      '&:hover': {
        backgroundColor: alpha(theme.palette.mode === 'dark'
          ? '#1a237e'
          : '#3949ab',
          theme.palette.action.hoverOpacity),
      },
    },
    '& .MuiSwitch-switchBase.Mui-checked + .MuiSwitch-track': {
      backgroundColor: theme.palette.mode === 'dark'
        ? '#1a237e'
        : '#3949ab',
    },
  },
  gradientText: {
    background: theme.palette.mode === 'dark'
      ? 'linear-gradient(45deg, #82b1ff 30%, #b388ff 90%)'
      : 'linear-gradient(45deg, #3949ab 30%, #5e35b1 90%)',
    WebkitBackgroundClip: 'text',
    WebkitTextFillColor: 'transparent',
    backgroundClip: 'text',
    textFillColor: 'transparent',
  },
  circularProgress: {
    background: theme.palette.mode === 'dark'
      ? 'linear-gradient(45deg, #1a237e 30%, #311b92 90%)'
      : 'linear-gradient(45deg, #3949ab 30%, #5e35b1 90%)',
    borderRadius: '50%',
    padding: '2px',
  },
  divider: {
    opacity: theme.palette.mode === 'dark' ? 0.1 : 0.2,
  },
  scrollbar: {
    '&::-webkit-scrollbar': {
      width: '8px',
      height: '8px',
    },
    '&::-webkit-scrollbar-track': {
      background: 'transparent',
    },
    '&::-webkit-scrollbar-thumb': {
      background: theme.palette.mode === 'dark'
        ? alpha(theme.palette.primary.main, 0.2)
        : alpha(theme.palette.primary.main, 0.1),
      borderRadius: '4px',
      '&:hover': {
        background: theme.palette.mode === 'dark'
          ? alpha(theme.palette.primary.main, 0.3)
          : alpha(theme.palette.primary.main, 0.2),
      }
    }
  }
}); 