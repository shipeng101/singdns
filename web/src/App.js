import React, { useState, useEffect } from 'react';
import {
  Box,
  CssBaseline,
  ThemeProvider,
  Drawer,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  IconButton,
  useMediaQuery,
  Typography,
  Container,
  Menu,
  MenuItem,
} from '@mui/material';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation, Link } from 'react-router-dom';
import getTheme from './theme';
import ErrorBoundary from './components/ErrorBoundary';

// Import your components
import Dashboard from './pages/Dashboard';
import Nodes from './pages/Nodes';
import Rules from './pages/Rules';
import Settings from './pages/Settings';
import Subscriptions from './pages/Subscriptions';

import {
  Dashboard as DashboardIcon,
  Dns as DnsIcon,
  Settings as SettingsIcon,
  Rule as RuleIcon,
  Menu as MenuIcon,
} from '@mui/icons-material';

const drawerWidth = 150;

const menuItems = [
  { text: '仪表盘', icon: <DashboardIcon />, path: '/' },
  { text: '节点管理', icon: <DnsIcon />, path: '/nodes' },
  { text: '规则管理', icon: <RuleIcon />, path: '/rules' },
  { text: '订阅管理', icon: <DnsIcon />, path: '/subscriptions' },
  { text: '系统设置', icon: <SettingsIcon />, path: '/settings' },
];

function App() {
  const [mode, setMode] = useState(() => {
    const savedMode = localStorage.getItem('themeMode');
    return savedMode || 'light';
  });
  const [mobileOpen, setMobileOpen] = useState(false);
  const [dashboardAnchorEl, setDashboardAnchorEl] = useState(null);
  const theme = getTheme(mode);

  useEffect(() => {
    localStorage.setItem('themeMode', mode);
  }, [mode]);

  return (
    <ThemeProvider theme={theme}>
      <ErrorBoundary>
        <Router>
          <Routes>
            <Route
              path="/*"
              element={
                <AppContent 
                  mode={mode} 
                  setMode={setMode}
                  mobileOpen={mobileOpen}
                  setMobileOpen={setMobileOpen}
                  dashboardAnchorEl={dashboardAnchorEl}
                  setDashboardAnchorEl={setDashboardAnchorEl}
                />
              }
            />
          </Routes>
        </Router>
      </ErrorBoundary>
    </ThemeProvider>
  );
}

function AppContent({ mode, setMode, mobileOpen, setMobileOpen, dashboardAnchorEl, setDashboardAnchorEl }) {
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'));
  const location = useLocation();

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const handleDashboardClick = (event) => {
    setDashboardAnchorEl(event.currentTarget);
  };

  const handleDashboardClose = () => {
    setDashboardAnchorEl(null);
  };

  const handleDashboardSelect = (url) => {
    window.open(url, '_blank');
    handleDashboardClose();
  };

  const drawer = (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <Box sx={{ 
        p: 2.5,
        display: 'flex', 
        alignItems: 'center', 
        justifyContent: 'center',
      }}>
        <Typography 
          variant="h6" 
          noWrap 
          component="div"
          sx={{ 
            fontSize: '1.4rem',
            fontWeight: 600,
            background: (theme) => 
              theme.palette.mode === 'dark' 
                ? 'linear-gradient(90deg, #90caf9 30%, #ce93d8 90%)'
                : 'linear-gradient(90deg, #1976d2 30%, #9c27b0 90%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            letterSpacing: '0.5px',
          }}
        >
          singdns
        </Typography>
      </Box>
      <List sx={{ 
        mt: 1,
        px: 1.5,
        py: 1,
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
                background: (theme) => 
                  theme.palette.mode === 'dark'
                    ? 'linear-gradient(90deg, rgba(144, 202, 249, 0.15) 0%, rgba(206, 147, 216, 0.15) 100%)'
                    : 'linear-gradient(90deg, rgba(25, 118, 210, 0.08) 0%, rgba(156, 39, 176, 0.08) 100%)',
                '&:hover': {
                  background: (theme) => 
                    theme.palette.mode === 'dark'
                      ? 'linear-gradient(90deg, rgba(144, 202, 249, 0.25) 0%, rgba(206, 147, 216, 0.25) 100%)'
                      : 'linear-gradient(90deg, rgba(25, 118, 210, 0.12) 0%, rgba(156, 39, 176, 0.12) 100%)',
                },
              },
              '&:hover': {
                background: (theme) => 
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
                    ? theme.palette.mode === 'dark'
                      ? '#90caf9'
                      : theme.palette.primary.main
                    : theme.palette.mode === 'dark'
                      ? 'rgba(255, 255, 255, 0.6)'
                      : 'rgba(0, 0, 0, 0.4)',
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
                      ? theme.palette.mode === 'dark'
                        ? '#fff'
                        : theme.palette.text.primary
                      : theme.palette.mode === 'dark'
                        ? 'rgba(255, 255, 255, 0.6)'
                        : 'rgba(0, 0, 0, 0.6)',
                }
              }}
            />
          </ListItem>
        ))}
      </List>
    </Box>
  );

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <CssBaseline />
      
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
              background: (theme) => 
                theme.palette.mode === 'dark' 
                  ? '#0a1929'
                  : '#f8fafc',
              borderRight: 'none',
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
          width: { sm: `calc(100% - ${drawerWidth}px)` },
          ml: { sm: `${drawerWidth}px` },
          backgroundColor: (theme) => theme.palette.background.default,
          minHeight: '100vh',
        }}
      >
        <Container 
          maxWidth={false} 
          sx={{ 
            py: 1,
            px: { xs: 1, sm: 1.5 },
            height: '100%',
          }}
        >
          <Menu
            anchorEl={dashboardAnchorEl}
            open={Boolean(dashboardAnchorEl)}
            onClose={handleDashboardClose}
            PaperProps={{
              elevation: 0,
              sx: {
                border: '1px solid',
                borderColor: (theme) => 
                  theme.palette.mode === 'dark' 
                    ? 'rgba(255, 255, 255, 0.12)'
                    : 'rgba(0, 0, 0, 0.06)',
              },
            }}
          >
            <MenuItem onClick={() => handleDashboardSelect('/ui/metacubexd/')}>
              MetaCubeXD
            </MenuItem>
            <MenuItem onClick={() => handleDashboardSelect('/ui/yacd/')}>
              YACD
            </MenuItem>
          </Menu>
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
        </Container>
      </Box>
    </Box>
  );
}

export default App; 