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
  Divider,
  Container,
  Tooltip,
  Menu,
  MenuItem
} from '@mui/material';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation, Link } from 'react-router-dom';
import getTheme from './theme';
import { initAuth, getToken } from './services/auth';
import Login from './pages/Login';

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
  DarkMode as DarkModeIcon,
  LightMode as LightModeIcon,
  Menu as MenuIcon,
  OpenInNew as OpenInNewIcon
} from '@mui/icons-material';

const drawerWidth = 180;

const menuItems = [
  { text: '仪表盘', icon: <DashboardIcon />, path: '/' },
  { text: '节点管理', icon: <DnsIcon />, path: '/nodes' },
  { text: '规则管理', icon: <RuleIcon />, path: '/rules' },
  { text: '订阅管理', icon: <DnsIcon />, path: '/subscriptions' },
  { text: '系统设置', icon: <SettingsIcon />, path: '/settings' },
];

const ProtectedRoute = ({ children }) => {
  const token = getToken();
  const location = useLocation();

  if (!token) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return children;
};

function App() {
  const [mode, setMode] = useState('light');
  const [mobileOpen, setMobileOpen] = useState(false);
  const [dashboardAnchorEl, setDashboardAnchorEl] = useState(null);
  const theme = getTheme(mode);

  useEffect(() => {
    initAuth();
  }, []);

  return (
    <ThemeProvider theme={theme}>
      <Router>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route
            path="/*"
            element={
              <ProtectedRoute>
                <AppContent 
                  mode={mode} 
                  setMode={setMode}
                  mobileOpen={mobileOpen}
                  setMobileOpen={setMobileOpen}
                  dashboardAnchorEl={dashboardAnchorEl}
                  setDashboardAnchorEl={setDashboardAnchorEl}
                />
              </ProtectedRoute>
            }
          />
        </Routes>
      </Router>
    </ThemeProvider>
  );
}

function AppContent({ mode, setMode, mobileOpen, setMobileOpen, dashboardAnchorEl, setDashboardAnchorEl }) {
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'));

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
    const baseUrl = window.location.origin;
    window.open(`${baseUrl}${url}`, '_blank');
    handleDashboardClose();
  };

  const drawer = (
    <Box>
      <Box sx={{ p: 2, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Typography variant="h6" noWrap component="div">
          SingDNS
        </Typography>
      </Box>
      <Divider />
      <List>
        {menuItems.map((item) => (
          <ListItem
            button
            key={item.text}
            component={Link}
            to={item.path}
            sx={{
              '&.Mui-selected': {
                backgroundColor: 'action.selected',
              },
            }}
          >
            <ListItemIcon>{item.icon}</ListItemIcon>
            <ListItemText primary={item.text} />
          </ListItem>
        ))}
      </List>
    </Box>
  );

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <CssBaseline />
      
      {/* App Bar */}
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
            },
          }}
        >
          {drawer}
        </Drawer>
      </Box>

      {/* Main content */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: { xs: 1, sm: 2 },
          width: { sm: `calc(100% - ${drawerWidth}px)` },
          ml: { sm: `${drawerWidth}px` },
        }}
      >
        <Container maxWidth="xl" sx={{ py: 2 }}>
          <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 2 }}>
            <Tooltip title="仪表盘">
              <IconButton onClick={handleDashboardClick} sx={{ mr: 1 }}>
                <OpenInNewIcon />
              </IconButton>
            </Tooltip>
            <Menu
              anchorEl={dashboardAnchorEl}
              open={Boolean(dashboardAnchorEl)}
              onClose={handleDashboardClose}
            >
              <MenuItem onClick={() => handleDashboardSelect('/ui/metacubexd/')}>
                MetaCubeXD
              </MenuItem>
              <MenuItem onClick={() => handleDashboardSelect('/ui/yacd/')}>
                YACD
              </MenuItem>
            </Menu>
            <Tooltip title={mode === 'dark' ? '切换亮色主题' : '切换暗色主题'}>
              <IconButton onClick={() => setMode(mode === 'dark' ? 'light' : 'dark')}>
                {mode === 'dark' ? <LightModeIcon /> : <DarkModeIcon />}
              </IconButton>
            </Tooltip>
          </Box>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/nodes" element={<Nodes />} />
            <Route path="/rules" element={<Rules />} />
            <Route path="/subscriptions" element={<Subscriptions />} />
            <Route path="/settings" element={<Settings />} />
          </Routes>
        </Container>
      </Box>
    </Box>
  );
}

export default App; 