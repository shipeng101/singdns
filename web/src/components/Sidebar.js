import React from 'react';
import {
  Drawer,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  useTheme,
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  Language as NodesIcon,
  Rule as RulesIcon,
  Subscriptions as SubscriptionsIcon,
  Settings as SettingsIcon,
} from '@mui/icons-material';
import { useLocation, useNavigate } from 'react-router-dom';

const menuItems = [
  { text: '仪表盘', icon: DashboardIcon, path: '/' },
  { text: '节点管理', icon: NodesIcon, path: '/nodes' },
  { text: '规则管理', icon: RulesIcon, path: '/rules' },
  { text: '订阅管理', icon: SubscriptionsIcon, path: '/subscriptions' },
  { text: '系统设置', icon: SettingsIcon, path: '/settings' },
];

const Sidebar = () => {
  const theme = useTheme();
  const location = useLocation();
  const navigate = useNavigate();

  return (
    <Drawer
      variant="permanent"
      sx={{
        width: 200,
        flexShrink: 0,
        '& .MuiDrawer-paper': {
          width: 200,
          boxSizing: 'border-box',
          backgroundColor: theme.palette.background.paper,
          borderRight: `1px solid ${theme.palette.divider}`,
        },
      }}
    >
      <List sx={{ mt: 8 }}>
        {menuItems.map((item) => (
          <ListItem
            button
            key={item.text}
            onClick={() => navigate(item.path)}
            selected={location.pathname === item.path}
            sx={{
              '&.Mui-selected': {
                backgroundColor: theme.palette.mode === 'dark'
                  ? 'rgba(144, 202, 249, 0.16)'
                  : 'rgba(25, 118, 210, 0.08)',
                '&:hover': {
                  backgroundColor: theme.palette.mode === 'dark'
                    ? 'rgba(144, 202, 249, 0.24)'
                    : 'rgba(25, 118, 210, 0.12)',
                },
              },
              borderRadius: 1,
              mx: 1,
              mb: 0.5,
            }}
          >
            <ListItemIcon sx={{ minWidth: 40 }}>
              <item.icon color={location.pathname === item.path ? 'primary' : 'inherit'} />
            </ListItemIcon>
            <ListItemText 
              primary={item.text}
              primaryTypographyProps={{
                color: location.pathname === item.path ? 'primary' : 'inherit',
                fontWeight: location.pathname === item.path ? 'bold' : 'normal',
              }}
            />
          </ListItem>
        ))}
      </List>
    </Drawer>
  );
};

export default Sidebar; 