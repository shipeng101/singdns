import React from 'react';
import {
  AppBar,
  Toolbar,
  IconButton,
  Box,
  useTheme,
  Tooltip,
} from '@mui/material';
import {
  Brightness4 as DarkIcon,
  Brightness7 as LightIcon,
  Notifications as NotificationsIcon,
  Settings as SettingsIcon,
} from '@mui/icons-material';
import Logo from './Logo';

const Header = ({ onToggleTheme }) => {
  const theme = useTheme();

  return (
    <AppBar
      position="fixed"
      sx={{
        zIndex: (theme) => theme.zIndex.drawer + 1,
        backgroundColor: theme.palette.background.paper,
        color: theme.palette.text.primary,
        boxShadow: 'none',
        borderBottom: `1px solid ${theme.palette.divider}`,
      }}
    >
      <Toolbar>
        <Logo />
        <Box sx={{ flexGrow: 1 }} />
        <Tooltip title={theme.palette.mode === 'dark' ? '切换到亮色模式' : '切换到暗色模式'}>
          <IconButton onClick={onToggleTheme} color="inherit">
            {theme.palette.mode === 'dark' ? <LightIcon /> : <DarkIcon />}
          </IconButton>
        </Tooltip>
        <Tooltip title="通知">
          <IconButton color="inherit">
            <NotificationsIcon />
          </IconButton>
        </Tooltip>
        <Tooltip title="设置">
          <IconButton color="inherit">
            <SettingsIcon />
          </IconButton>
        </Tooltip>
      </Toolbar>
    </AppBar>
  );
};

export default Header; 