import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Grid,
  Button,
  Typography,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Stack,
  Alert,
  Snackbar,
  CircularProgress,
  LinearProgress,
  useTheme,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction
} from '@mui/material';
import {
  DarkMode as DarkModeIcon,
  LightMode as LightModeIcon,
  OpenInNew as OpenInNewIcon,
  Add as AddIcon,
  Delete as DeleteIcon
} from '@mui/icons-material';
import { getSettings, updateSettings } from '../services/api';
import { getCommonStyles } from '../styles/commonStyles';

const Settings = ({ mode, setMode, onDashboardClick }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  
  // Password Management State
  const [passwordData, setPasswordData] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  });

  // DNS Settings State
  const [dnsSettings, setDnsSettings] = useState({
    domestic: [],
    foreign: [],
    currentDomestic: '',
    currentForeign: ''
  });
  const [dnsDialog, setDnsDialog] = useState(false);
  const [dnsType, setDnsType] = useState('domestic');
  const [newDns, setNewDns] = useState({
    type: 'udp',
    address: ''
  });

  // Network Settings State
  const [networkSettings, setNetworkSettings] = useState({
    listenAddress: '',
    listenPort: '',
    currentAddress: '',
    currentPort: ''
  });

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      setLoading(true);
      const settings = await getSettings();
      
      // Update DNS settings
      setDnsSettings({
        domestic: settings.dns.domestic || [],
        foreign: settings.dns.foreign || [],
        currentDomestic: settings.dns.currentDomestic || '',
        currentForeign: settings.dns.currentForeign || ''
      });

      // Update Network settings
      setNetworkSettings({
        listenAddress: settings.network.listenAddress || '',
        listenPort: settings.network.listenPort || '',
        currentAddress: settings.network.currentAddress || '',
        currentPort: settings.network.currentPort || ''
      });

    } catch (err) {
      setError(err.response?.data?.error || '获取设置失败');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordChange = async () => {
    if (passwordData.newPassword !== passwordData.confirmPassword) {
      setError('新密码与确认密码不匹配');
      return;
    }

    try {
      setLoading(true);
      await updateSettings({
        password: {
          current: passwordData.currentPassword,
          new: passwordData.newPassword
        }
      });
      setSuccess(true);
      setPasswordData({
        currentPassword: '',
        newPassword: '',
        confirmPassword: ''
      });
    } catch (err) {
      setError(err.response?.data?.error || '修改密码失败');
    } finally {
      setLoading(false);
    }
  };

  const handleAddDns = async () => {
    try {
      setLoading(true);
      const updatedSettings = {
        dns: {
          ...dnsSettings,
          [dnsType]: [...dnsSettings[dnsType], newDns]
        }
      };
      await updateSettings(updatedSettings);
      setDnsSettings(updatedSettings.dns);
      setDnsDialog(false);
      setNewDns({ type: 'udp', address: '' });
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '添加DNS失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteDns = async (type, index) => {
    try {
      setLoading(true);
      const updatedDns = [...dnsSettings[type]];
      updatedDns.splice(index, 1);
      const updatedSettings = {
        dns: {
          ...dnsSettings,
          [type]: updatedDns
        }
      };
      await updateSettings(updatedSettings);
      setDnsSettings(updatedSettings.dns);
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '删除DNS失败');
    } finally {
      setLoading(false);
    }
  };

  const handleNetworkSettingsUpdate = async () => {
    try {
      setLoading(true);
      await updateSettings({
        network: {
          listenAddress: networkSettings.listenAddress,
          listenPort: networkSettings.listenPort
        }
      });
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '更新网络设置失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box>
      <Snackbar 
        open={success} 
        autoHideDuration={3000} 
        onClose={() => setSuccess(false)}
      >
        <Alert severity="success">操作成功</Alert>
      </Snackbar>

      <Snackbar 
        open={!!error} 
        autoHideDuration={3000} 
        onClose={() => setError(null)}
      >
        <Alert severity="error">{error}</Alert>
      </Snackbar>

      <Card sx={styles.headerCard}>
        <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Stack direction="row" alignItems="center" spacing={2}>
              <Typography variant="h5" sx={{ fontWeight: 500, color: 'inherit' }}>
                系统设置
              </Typography>
              {loading && (
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <CircularProgress size={20} thickness={4} sx={{ color: 'rgba(255,255,255,0.8)' }} />
                </Box>
              )}
            </Stack>
            <Stack direction="row" spacing={1}>
              <IconButton
                size="small"
                onClick={onDashboardClick}
                sx={styles.iconButton}
              >
                <OpenInNewIcon fontSize="small" />
              </IconButton>
              <IconButton
                size="small"
                onClick={() => setMode(mode === 'dark' ? 'light' : 'dark')}
                sx={styles.iconButton}
              >
                {mode === 'dark' ? <LightModeIcon fontSize="small" /> : <DarkModeIcon fontSize="small" />}
              </IconButton>
            </Stack>
          </Stack>
        </CardContent>
      </Card>

      {loading && <LinearProgress sx={{ mb: 1 }} />}

      <Grid container spacing={1.5} sx={{ mt: 0.5 }}>
        {/* Password Management Card */}
        <Grid item xs={12} md={6}>
          <Card sx={styles.card}>
            <CardContent>
              <Typography variant="h6" gutterBottom>密码管理</Typography>
              <Stack spacing={2}>
                <TextField
                  fullWidth
                  type="password"
                  label="当前密码"
                  value={passwordData.currentPassword}
                  onChange={(e) => setPasswordData({
                    ...passwordData,
                    currentPassword: e.target.value
                  })}
                  size="small"
                />
                <TextField
                  fullWidth
                  type="password"
                  label="新密码"
                  value={passwordData.newPassword}
                  onChange={(e) => setPasswordData({
                    ...passwordData,
                    newPassword: e.target.value
                  })}
                  size="small"
                />
                <TextField
                  fullWidth
                  type="password"
                  label="确认新密码"
                  value={passwordData.confirmPassword}
                  onChange={(e) => setPasswordData({
                    ...passwordData,
                    confirmPassword: e.target.value
                  })}
                  size="small"
                />
                <Button
                  variant="contained"
                  size="small"
                  onClick={handlePasswordChange}
                  sx={styles.actionButton}
                >
                  修改密码
                </Button>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        {/* Network Settings Card */}
        <Grid item xs={12} md={6}>
          <Card sx={styles.card}>
            <CardContent>
              <Typography variant="h6" gutterBottom>网络设置</Typography>
              <Stack spacing={2}>
                <TextField
                  fullWidth
                  size="small"
                  label="监听地址"
                  value={networkSettings.listenAddress}
                  onChange={(e) => setNetworkSettings({
                    ...networkSettings,
                    listenAddress: e.target.value
                  })}
                  placeholder={networkSettings.currentAddress}
                  helperText={`当前地址: ${networkSettings.currentAddress}`}
                />
                <TextField
                  fullWidth
                  size="small"
                  label="监听端口"
                  value={networkSettings.listenPort}
                  onChange={(e) => setNetworkSettings({
                    ...networkSettings,
                    listenPort: e.target.value
                  })}
                  placeholder={networkSettings.currentPort}
                  helperText={`当前端口: ${networkSettings.currentPort}`}
                />
                <Button
                  variant="contained"
                  size="small"
                  onClick={handleNetworkSettingsUpdate}
                  sx={styles.actionButton}
                >
                  保存设置
                </Button>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        {/* DNS Settings Cards */}
        <Grid item xs={12} md={6}>
          <Card sx={styles.card}>
            <CardContent>
              <Stack direction="row" justifyContent="space-between" alignItems="center" mb={2}>
                <Typography variant="h6">国内DNS设置</Typography>
                <Button
                  variant="contained"
                  size="small"
                  startIcon={<AddIcon />}
                  onClick={() => {
                    setDnsType('domestic');
                    setDnsDialog(true);
                  }}
                  sx={styles.actionButton}
                >
                  添加DNS
                </Button>
              </Stack>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                当前使用: {dnsSettings.currentDomestic || '未设置'}
              </Typography>
              <List>
                {dnsSettings.domestic.map((dns, index) => (
                  <ListItem key={index}>
                    <ListItemText
                      primary={dns.address}
                      secondary={dns.type.toUpperCase()}
                    />
                    <ListItemSecondaryAction>
                      <IconButton
                        edge="end"
                        size="small"
                        onClick={() => handleDeleteDns('domestic', index)}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </ListItemSecondaryAction>
                  </ListItem>
                ))}
              </List>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Card sx={styles.card}>
            <CardContent>
              <Stack direction="row" justifyContent="space-between" alignItems="center" mb={2}>
                <Typography variant="h6">国外DNS设置</Typography>
                <Button
                  variant="contained"
                  size="small"
                  startIcon={<AddIcon />}
                  onClick={() => {
                    setDnsType('foreign');
                    setDnsDialog(true);
                  }}
                  sx={styles.actionButton}
                >
                  添加DNS
                </Button>
              </Stack>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                当前使用: {dnsSettings.currentForeign || '未设置'}
              </Typography>
              <List>
                {dnsSettings.foreign.map((dns, index) => (
                  <ListItem key={index}>
                    <ListItemText
                      primary={dns.address}
                      secondary={dns.type.toUpperCase()}
                    />
                    <ListItemSecondaryAction>
                      <IconButton
                        edge="end"
                        size="small"
                        onClick={() => handleDeleteDns('foreign', index)}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </ListItemSecondaryAction>
                  </ListItem>
                ))}
              </List>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Add DNS Dialog */}
      <Dialog 
        open={dnsDialog} 
        onClose={() => setDnsDialog(false)}
        maxWidth="sm"
        fullWidth
        PaperProps={{
          sx: styles.dialog
        }}
      >
        <DialogTitle>
          添加{dnsType === 'domestic' ? '国内' : '国外'}DNS
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 2 }}>
            <FormControl fullWidth size="small">
              <InputLabel>DNS类型</InputLabel>
              <Select
                value={newDns.type}
                label="DNS类型"
                onChange={(e) => setNewDns({
                  ...newDns,
                  type: e.target.value
                })}
              >
                <MenuItem value="udp">UDP</MenuItem>
                <MenuItem value="tcp">TCP</MenuItem>
                <MenuItem value="doh">DoH</MenuItem>
                <MenuItem value="dot">DoT</MenuItem>
              </Select>
            </FormControl>
            <TextField
              fullWidth
              label="DNS地址"
              value={newDns.address}
              onChange={(e) => setNewDns({
                ...newDns,
                address: e.target.value
              })}
              placeholder={newDns.type === 'doh' ? 'https://dns.example.com/dns-query' : '8.8.8.8:53'}
              size="small"
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDnsDialog(false)}>取消</Button>
          <Button variant="contained" onClick={handleAddDns}>
            添加
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Settings; 