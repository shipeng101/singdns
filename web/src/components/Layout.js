import React, { useState, useEffect } from 'react';
import {
  Box,
  Drawer,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Typography,
  IconButton,
  useTheme,
  useMediaQuery,
} from '@mui/material';
import {
  Routes,
  Route,
  Navigate,
  useLocation,
  Link,
} from 'react-router-dom';
import MenuIcon from '@mui/icons-material/Menu';
import DashboardIcon from '@mui/icons-material/Dashboard';
import DeviceHubIcon from '@mui/icons-material/DeviceHub';
import RuleIcon from '@mui/icons-material/Rule';
import SubscriptionsIcon from '@mui/icons-material/Subscriptions';
import SettingsIcon from '@mui/icons-material/Settings';
import ErrorBoundary from './ErrorBoundary';
import ProtectedRoute from './ProtectedRoute';
import Login from '../pages/Login';
import Dashboard from '../pages/Dashboard';
import Nodes from '../pages/Nodes';
import Rules from '../pages/Rules';
import Subscriptions from '../pages/Subscriptions';
import Settings from '../pages/Settings';

const drawerWidth = 200;

const menuItems = [
  { text: '仪表盘', path: '/', icon: <DashboardIcon /> },
  { text: '节点管理', path: '/nodes', icon: <DeviceHubIcon /> },
  { text: '规则管理', path: '/rules', icon: <RuleIcon /> },
  { text: '订阅管理', path: '/subscriptions', icon: <SubscriptionsIcon /> },
  { text: '系统设置', path: '/settings', icon: <SettingsIcon /> },
];

function Layout({ mode, setMode }) {
  const theme = useTheme();
  const location = useLocation();
  const [mobileOpen, setMobileOpen] = useState(false);
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  // 处理系统主题变化
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = (e) => {
      const savedMode = localStorage.getItem('theme_mode');
      if (savedMode === 'system') {
        setMode(e.matches ? 'dark' : 'light');
      }
    };

    // 初始化时检查
    const savedMode = localStorage.getItem('theme_mode');
    if (!savedMode || savedMode === 'system') {
      setMode(mediaQuery.matches ? 'dark' : 'light');
    }

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, [setMode]);

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const handleDashboardClick = () => {
    const protocol = window.location.protocol;
    const hostname = window.location.hostname;
    const dashboardUrl = `${protocol}//${hostname}:9000/ui/`;
    window.open(dashboardUrl, '_blank');
  };

  // 检查是否在登录页面
  const isLoginPage = location.pathname === '/login';

  // 如果是登录页面，直接渲染登录组件
  if (isLoginPage) {
    return <Login />;
  }

  const drawer = (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <Box sx={{ 
        p: 2,
        display: 'flex', 
        alignItems: 'center', 
        justifyContent: 'center',
      }}>
        <Typography 
          variant="h6" 
          noWrap 
          component="div"
          sx={{ 
            fontSize: '1.2rem',
            fontWeight: 600,
            color: (theme) => 
              theme.palette.mode === 'dark' 
                ? '#90caf9'
                : '#1976d2',
          }}
        >
          singdns
        </Typography>
      </Box>
      <List sx={{ 
        mt: 0.5,
        px: 1,
        py: 0.5,
        flexGrow: 1,
      }}>
        {menuItems.map((item) => (
          <ListItem
            button
            key={item.text}
            component={Link}
            to={item.path}
            selected={location.pathname === item.path}
            onClick={() => isMobile && handleDrawerToggle()}
            sx={{
              minHeight: 44,
              mb: 0.5,
              borderRadius: '12px',
              px: 1.5,
              transition: 'all 0.2s',
              '&.Mui-selected': {
                bgcolor: (theme) => 
                  theme.palette.mode === 'dark'
                    ? theme.palette.primary.dark
                    : theme.palette.primary.light,
                '&:hover': {
                  bgcolor: (theme) => 
                    theme.palette.mode === 'dark'
                      ? theme.palette.primary.main
                      : theme.palette.primary.light,
                },
              },
              '&:hover': {
                bgcolor: (theme) => 
                  theme.palette.mode === 'dark'
                    ? 'rgba(255, 255, 255, 0.05)'
                    : 'rgba(0, 0, 0, 0.02)',
              },
            }}
          >
            <ListItemIcon 
              sx={{ 
                minWidth: 36,
                transition: 'color 0.2s',
                color: (theme) => 
                  location.pathname === item.path
                    ? theme.palette.primary.main
                    : theme.palette.text.secondary,
              }}
            >
              {item.icon}
            </ListItemIcon>
            <ListItemText 
              primary={item.text}
              primaryTypographyProps={{
                fontSize: '0.9rem',
                fontWeight: location.pathname === item.path ? 600 : 400,
                sx: {
                  transition: 'color 0.2s',
                  color: (theme) => 
                    location.pathname === item.path
                      ? theme.palette.text.primary
                      : theme.palette.text.secondary,
                }
              }}
            />
          </ListItem>
        ))}
      </List>
    </Box>
  );

  return (
    <ErrorBoundary>
      <Box sx={{ display: 'flex', minHeight: '100vh' }}>
        <Box
          component="nav"
          sx={{
            width: { sm: drawerWidth },
            flexShrink: { sm: 0 }
          }}
        >
          {isMobile && (
            <IconButton
              color="inherit"
              aria-label="open drawer"
              edge="start"
              onClick={handleDrawerToggle}
              sx={{ mr: 2, display: { sm: 'none' } }}
            >
              <MenuIcon />
            </IconButton>
          )}
          <Drawer
            variant={isMobile ? 'temporary' : 'permanent'}
            open={isMobile ? mobileOpen : true}
            onClose={handleDrawerToggle}
            ModalProps={{
              keepMounted: true,
            }}
            sx={{
              '& .MuiDrawer-paper': {
                boxSizing: 'border-box',
                width: drawerWidth,
                background: (theme) => theme.palette.background.paper,
                borderRight: (theme) =>
                  `1px solid ${theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.12)' : 'rgba(0, 0, 0, 0.12)'}`,
              },
            }}
          >
            {drawer}
          </Drawer>
        </Box>

        <Box
          component="main"
          sx={{
            flexGrow: 1,
            width: { sm: `calc(100% - ${drawerWidth + 160}px)` },
            ml: { sm: `${drawerWidth - 100}px` },
            mr: { sm: '60px' },
            backgroundColor: (theme) => theme.palette.background.default,
            minHeight: '100vh',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <Box 
            sx={{ 
              flexGrow: 1,
              py: 1,
              pl: 0,
              pr: 0,
              height: '100%',
            }}
          >
            <Routes>
              <Route path="/login" element={<Login />} />
              <Route
                path="/*"
                element={
                  <ProtectedRoute>
                    <Routes>
                      <Route path="/" element={
                        <Dashboard 
                          mode={mode} 
                          setMode={setMode}
                          onDashboardClick={handleDashboardClick}
                        />
                      } />
                      <Route path="/nodes" element={
                        <Nodes 
                          mode={mode} 
                          setMode={setMode}
                          onDashboardClick={handleDashboardClick}
                        />
                      } />
                      <Route path="/rules" element={
                        <Rules 
                          mode={mode} 
                          setMode={setMode}
                          onDashboardClick={handleDashboardClick}
                        />
                      } />
                      <Route path="/subscriptions" element={
                        <Subscriptions 
                          mode={mode} 
                          setMode={setMode}
                          onDashboardClick={handleDashboardClick}
                        />
                      } />
                      <Route path="/settings" element={
                        <Settings 
                          mode={mode} 
                          setMode={setMode}
                          onDashboardClick={handleDashboardClick}
                        />
                      } />
                      <Route path="*" element={<Navigate to="/" replace />} />
                    </Routes>
                  </ProtectedRoute>
                }
              />
            </Routes>
          </Box>
        </Box>
      </Box>
    </ErrorBoundary>
  );
}

export default Layout; 