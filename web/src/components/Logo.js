import React from 'react';
import { Typography, Box } from '@mui/material';
import { useTheme } from '@mui/material/styles';

const Logo = () => {
  const theme = useTheme();
  
  return (
    <Box sx={{ display: 'flex', alignItems: 'center', p: 2 }}>
      <Typography
        variant="h5"
        sx={{
          fontWeight: 600,
          backgroundImage: theme.palette.mode === 'dark'
            ? 'linear-gradient(45deg, #90caf9 30%, #ce93d8 90%)'
            : 'linear-gradient(45deg, #1976d2 30%, #9c27b0 90%)',
          WebkitBackgroundClip: 'text',
          WebkitTextFillColor: 'transparent',
          backgroundClip: 'text',
          textFillColor: 'transparent',
        }}
      >
        singdns
      </Typography>
    </Box>
  );
};

export default Logo; 