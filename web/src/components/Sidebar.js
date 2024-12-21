import React from 'react';
import {
  Box,
  Drawer,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Typography,
  useTheme
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  Rule as RuleIcon,
  Router as RouterIcon,
  Subscriptions as SubscriptionsIcon,
  Settings as SettingsIcon
} from '@mui/icons-material';
import { useLocation, useNavigate } from 'react-router-dom';

const Sidebar = ({ open, onClose }) => {
  const theme = useTheme();
  const location = useLocation();
  const navigate = useNavigate();

  const menuItems = [
    { text: '仪表盘', icon: <DashboardIcon />, path: '/' },
    { text: '节点管理', icon: <RouterIcon />, path: '/nodes' },
    { text: '规则管理', icon: <RuleIcon />, path: '/rules' },
    { text: '订阅管理', icon: <SubscriptionsIcon />, path: '/subscriptions' },
    { text: '系统设置', icon: <SettingsIcon />, path: '/settings' }
  ];

  const drawerWidth = 240;

  const isSelected = (path) => {
    if (path === '/') {
      return location.pathname === '/';
    }
    return location.pathname.startsWith(path);
  };

  return (
    <Drawer
      variant="permanent"
      open={open}
      onClose={onClose}
      sx={{
        width: drawerWidth,
        flexShrink: 0,
        '& .MuiDrawer-paper': {
          width: drawerWidth,
          boxSizing: 'border-box',
          background: theme.palette.mode === 'dark'
            ? 'linear-gradient(145deg, rgba(26,35,126,0.7), rgba(49,27,146,0.7))'
            : 'linear-gradient(145deg, rgba(255,255,255,0.9), rgba(255,255,255,0.95))',
          backdropFilter: 'blur(10px)',
          border: theme.palette.mode === 'dark'
            ? '1px solid rgba(255,255,255,0.1)'
            : '1px solid rgba(255,255,255,0.5)',
          boxShadow: theme.palette.mode === 'dark'
            ? '0 4px 12px 0 rgba(0,0,0,0.3), 0 2px 4px -1px rgba(0,0,0,0.4)'
            : '0 4px 12px 0 rgba(31,38,135,0.15), 0 2px 4px -1px rgba(31,38,135,0.2)',
        },
      }}
    >
      <Box sx={{ p: 2, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Typography 
          variant="h5" 
          sx={{ 
            fontWeight: 600,
            background: theme.palette.mode === 'dark'
              ? 'linear-gradient(45deg, #82b1ff 0%, #b388ff 100%)'
              : 'linear-gradient(45deg, #3949ab 0%, #5e35b1 100%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            textAlign: 'center',
            filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.2))',
          }}
        >
          singdns
        </Typography>
      </Box>
      <List sx={{ px: 1 }}>
        {menuItems.map((item) => (
          <ListItem key={item.text} disablePadding sx={{ mb: 0.5 }}>
            <ListItemButton
              selected={isSelected(item.path)}
              onClick={() => navigate(item.path)}
              sx={{
                borderRadius: 1,
                color: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.7)' : 'rgba(0,0,0,0.7)',
                '&.Mui-selected': {
                  background: theme.palette.mode === 'dark'
                    ? 'linear-gradient(90deg, rgba(255,255,255,0.1), rgba(255,255,255,0.05))'
                    : 'linear-gradient(90deg, rgba(57,73,171,0.1), rgba(94,53,177,0.05))',
                  color: theme.palette.mode === 'dark' ? '#fff' : '#3949ab',
                  '&:hover': {
                    background: theme.palette.mode === 'dark'
                      ? 'linear-gradient(90deg, rgba(255,255,255,0.15), rgba(255,255,255,0.1))'
                      : 'linear-gradient(90deg, rgba(57,73,171,0.15), rgba(94,53,177,0.1))',
                  },
                  '& .MuiListItemIcon-root': {
                    color: 'inherit',
                  },
                },
                '&:hover': {
                  background: theme.palette.mode === 'dark'
                    ? 'linear-gradient(90deg, rgba(255,255,255,0.08), rgba(255,255,255,0.04))'
                    : 'linear-gradient(90deg, rgba(57,73,171,0.08), rgba(94,53,177,0.04))',
                  color: theme.palette.mode === 'dark' ? '#fff' : '#3949ab',
                  '& .MuiListItemIcon-root': {
                    color: 'inherit',
                  },
                },
              }}
            >
              <ListItemIcon 
                sx={{ 
                  minWidth: 40,
                  color: 'inherit',
                  transition: 'color 0.2s',
                }}
              >
                {item.icon}
              </ListItemIcon>
              <ListItemText 
                primary={item.text} 
                primaryTypographyProps={{
                  sx: {
                    fontSize: '0.95rem',
                    fontWeight: isSelected(item.path) ? 500 : 400,
                  }
                }}
              />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
    </Drawer>
  );
};

export default Sidebar; 