import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
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
  FormControlLabel,
  Switch,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Chip,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  ListItemIcon,
  Divider,
  Grid
} from '@mui/material';
import {
  DarkMode as DarkModeIcon,
  LightMode as LightModeIcon,
  OpenInNew as OpenInNewIcon,
  Add as AddIcon,
  Delete as DeleteIcon,
  Edit as EditIcon,
  Sync as SyncIcon,
  Dashboard as DashboardIcon,
} from '@mui/icons-material';
import { getSubscriptions, createSubscription, updateSubscription, deleteSubscription, refreshSubscription } from '../services/api';
import { getCommonStyles } from '../styles/commonStyles';

const Subscriptions = ({ mode, setMode, onDashboardClick }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  const [subscriptions, setSubscriptions] = useState([]);
  const [openDialog, setOpenDialog] = useState(false);
  const [selectedSubscription, setSelectedSubscription] = useState(null);
  const [editSubscription, setEditSubscription] = useState({
    id: '',
    name: '',
    url: '',
    type: 'v2ray',
    active: true,
    auto_update: true,
    update_interval: 3600,
  });

  const subscriptionTypes = [
    { value: 'singbox', label: 'SingBox', description: 'SingBox 订阅格式' },
    { value: 'clash', label: 'Clash', description: 'Clash 订阅格式' },
    { value: 'v2ray', label: 'V2Ray', description: 'V2Ray 订阅格式' },
    { value: 'shadowsocks', label: 'Shadowsocks', description: 'Shadowsocks 订阅格式' }
  ];

  const updateIntervalUnits = [
    { value: 'minute', label: '分钟' },
    { value: 'hour', label: '小时' },
    { value: 'day', label: '天' },
    { value: 'week', label: '周' }
  ];

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const data = await getSubscriptions();
      const processedData = data.map(subscription => ({
        ...subscription,
        active: subscription.active ?? true
      }));
      setSubscriptions(processedData);
      setError(null);
    } catch (err) {
      setError(err.response?.data?.error || '获取数据失败');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = (subscription = null) => {
    if (subscription) {
      setSelectedSubscription(subscription);
      setEditSubscription({
        ...subscription,
        updateInterval: parseUpdateInterval(subscription.updateInterval),
        active: subscription.active ?? true
      });
    } else {
      setSelectedSubscription(null);
      setEditSubscription({
        name: '',
        type: 'singbox',
        url: '',
        autoUpdate: false,
        updateInterval: {
          value: 1,
          unit: 'hour'
        },
        active: true
      });
    }
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setSelectedSubscription(null);
    setEditSubscription({
      name: '',
      type: 'singbox',
      url: '',
      autoUpdate: false,
      updateInterval: {
        value: 1,
        unit: 'hour'
      },
      active: true
    });
  };

  const parseUpdateInterval = (interval) => {
    if (!interval) return { value: 1, unit: 'hour' };
    const value = parseInt(interval);
    if (value % (7 * 24 * 60) === 0) return { value: value / (7 * 24 * 60), unit: 'week' };
    if (value % (24 * 60) === 0) return { value: value / (24 * 60), unit: 'day' };
    if (value % 60 === 0) return { value: value / 60, unit: 'hour' };
    return { value, unit: 'minute' };
  };

  const formatUpdateInterval = (interval) => {
    const { value, unit } = interval;
    switch (unit) {
      case 'week': return value * 7 * 24 * 60;
      case 'day': return value * 24 * 60;
      case 'hour': return value * 60;
      default: return value;
    }
  };

  const handleSave = async () => {
    try {
      const payload = {
        ...editSubscription,
        active: Boolean(editSubscription.active),
        auto_update: Boolean(editSubscription.auto_update),
        update_interval: Number(editSubscription.update_interval),
      };

      setOpenDialog(false);

      if (editSubscription.id) {
        await updateSubscription(editSubscription.id, payload);
      } else {
        await createSubscription(payload);
      }
      
      await fetchData();
      setSuccess(true);
    } catch (error) {
      console.error('Failed to save subscription:', error);
      setError(error.message || '保存失败');
    }
  };

  const handleDelete = async (id) => {
    try {
      setLoading(true);
      await deleteSubscription(id);
      await fetchData();
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '删除订阅失败');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdate = async (id) => {
    try {
      setLoading(true);
      await refreshSubscription(id);
      await fetchData();
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '更新订阅失败');
    } finally {
      setLoading(false);
    }
  };

  const handleManualRefresh = async () => {
    await fetchData();
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
                订阅管理
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
                onClick={handleManualRefresh}
                disabled={loading}
                sx={styles.iconButton}
                title="刷新"
              >
                <SyncIcon fontSize="small" />
              </IconButton>
              <IconButton
                size="small"
                onClick={onDashboardClick}
                sx={styles.iconButton}
                title="打开仪表盘"
              >
                <DashboardIcon fontSize="small" />
              </IconButton>
              <IconButton
                size="small"
                onClick={() => setMode(mode === 'dark' ? 'light' : 'dark')}
                sx={styles.iconButton}
                title={mode === 'dark' ? '切换到亮色模式' : '切换到暗色模式'}
              >
                {mode === 'dark' ? <LightModeIcon fontSize="small" /> : <DarkModeIcon fontSize="small" />}
              </IconButton>
            </Stack>
          </Stack>
        </CardContent>
      </Card>

      {loading && <LinearProgress sx={{ mb: 1 }} />}

      <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 2 }}>
        <Button
          variant="contained"
          size="small"
          startIcon={<AddIcon />}
          onClick={() => handleOpenDialog()}
          sx={styles.actionButton}
        >
          添加订阅
        </Button>
      </Box>

      <Grid container spacing={2} sx={{ maxWidth: '1600px', mx: 'auto' }}>
        {subscriptions.map((subscription) => (
          <Grid item xs={12} sm={6} md={4} lg={3} key={subscription.id}>
            <Box sx={{ maxWidth: '280px', mx: 'auto', width: '100%' }}>
              <Card sx={{
                position: 'relative',
                transition: 'transform 0.2s, box-shadow 0.2s',
                '&:hover': {
                  transform: 'translateY(-2px)',
                  boxShadow: (theme) => theme.shadows[4],
                },
                '&::before': {
                  content: '""',
                  display: 'block',
                  paddingTop: '75%',
                },
                '& > *': {
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  right: 0,
                  bottom: 0,
                  display: 'flex',
                  flexDirection: 'column',
                },
              }}>
                <CardContent sx={{ 
                  flexGrow: 1,
                  p: 1.5,
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'space-between',
                  '&:last-child': { pb: 1.5 },
                }}>
                  <Stack spacing={1}>
                    <Stack direction="row" spacing={0.5} flexWrap="wrap" gap={0.5} alignItems="center">
                      <Typography variant="subtitle1" sx={{ 
                        fontWeight: 500,
                        fontSize: '0.95rem',
                      }}>
                        {subscription.name}
                      </Typography>
                      <Chip 
                        label={`${subscription.node_count || 0} 个节点`}
                        size="small"
                        sx={{
                          bgcolor: '#2196f3',
                          color: '#fff',
                          fontWeight: 500,
                          height: '20px',
                          '& .MuiChip-label': {
                            px: 1,
                            fontSize: '0.75rem',
                          },
                        }}
                      />
                    </Stack>

                    <Stack direction="row" spacing={0.5} flexWrap="wrap" gap={0.5}>
                      <Chip 
                        label={subscription.type} 
                        size="small"
                        sx={{
                          bgcolor: '#3949ab',
                          color: '#fff',
                          fontWeight: 500,
                          height: '20px',
                          '& .MuiChip-label': {
                            px: 1,
                            fontSize: '0.75rem',
                          },
                        }}
                      />
                      <Chip 
                        label={subscription.active ? "已启用" : "已禁用"} 
                        size="small"
                        sx={subscription.active ? {
                          bgcolor: '#4caf50',
                          color: '#fff',
                          fontWeight: 500,
                          height: '20px',
                          '& .MuiChip-label': {
                            px: 1,
                            fontSize: '0.75rem',
                          },
                        } : {
                          bgcolor: '#f44336',
                          color: '#fff',
                          fontWeight: 500,
                          height: '20px',
                          '& .MuiChip-label': {
                            px: 1,
                            fontSize: '0.75rem',
                          },
                        }}
                      />
                      {subscription.autoUpdate && (
                        <Chip 
                          label={`${subscription.updateInterval}分钟更新`} 
                          size="small"
                          sx={{
                            bgcolor: '#ff9800',
                            color: '#fff',
                            fontWeight: 500,
                            height: '20px',
                            '& .MuiChip-label': {
                              px: 1,
                              fontSize: '0.75rem',
                            },
                          }}
                        />
                      )}
                    </Stack>

                    <Typography 
                      variant="body2" 
                      color="text.secondary" 
                      sx={{ 
                        wordBreak: 'break-all',
                        display: '-webkit-box',
                        WebkitLineClamp: 2,
                        WebkitBoxOrient: 'vertical',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        fontSize: '0.8rem',
                      }}
                    >
                      {subscription.url}
                    </Typography>
                  </Stack>

                  <Stack direction="row" spacing={1} justifyContent="flex-end">
                    <IconButton
                      size="small"
                      onClick={() => handleUpdate(subscription.id)}
                      sx={{
                        color: 'text.secondary',
                        padding: '4px',
                        '&:hover': {
                          color: 'primary.main',
                          backgroundColor: 'action.hover',
                        },
                      }}
                    >
                      <SyncIcon sx={{ fontSize: '1.1rem' }} />
                    </IconButton>
                    <IconButton
                      size="small"
                      onClick={() => handleOpenDialog(subscription)}
                      sx={{
                        color: 'text.secondary',
                        padding: '4px',
                        '&:hover': {
                          color: 'primary.main',
                          backgroundColor: 'action.hover',
                        },
                      }}
                    >
                      <EditIcon sx={{ fontSize: '1.1rem' }} />
                    </IconButton>
                    <IconButton
                      size="small"
                      onClick={() => handleDelete(subscription.id)}
                      sx={{
                        color: 'text.secondary',
                        padding: '4px',
                        '&:hover': {
                          color: 'error.main',
                          backgroundColor: 'action.hover',
                        },
                      }}
                    >
                      <DeleteIcon sx={{ fontSize: '1.1rem' }} />
                    </IconButton>
                  </Stack>
                </CardContent>
              </Card>
            </Box>
          </Grid>
        ))}
      </Grid>

      {/* Subscription Dialog */}
      <Dialog 
        open={openDialog} 
        onClose={handleCloseDialog}
        maxWidth="sm"
        fullWidth
        PaperProps={{
          sx: styles.dialog
        }}
      >
        <DialogTitle>
          {selectedSubscription ? '编辑订阅' : '添加订阅'}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 2 }}>
            <TextField
              fullWidth
              label="订阅名称"
              value={editSubscription.name}
              onChange={(e) => setEditSubscription({ ...editSubscription, name: e.target.value })}
              size="small"
            />
            <TextField
              fullWidth
              label="订阅地址"
              value={editSubscription.url}
              onChange={(e) => setEditSubscription({ ...editSubscription, url: e.target.value })}
              size="small"
              placeholder="https://example.com/subscribe"
              helperText="支持 Clash、V2Ray、Shadowsocks 等多种格式，将自动识别订阅类型"
            />
            <FormControlLabel
              control={
                <Switch
                  checked={editSubscription.active}
                  onChange={(e) => setEditSubscription({
                    ...editSubscription,
                    active: e.target.checked,
                  })}
                />
              }
              label="启用"
            />
            <FormControlLabel
              control={
                <Switch
                  checked={editSubscription.auto_update}
                  onChange={(e) => setEditSubscription({
                    ...editSubscription,
                    auto_update: e.target.checked,
                  })}
                />
              }
              label="自动更新"
            />
            {editSubscription.auto_update && (
              <TextField
                label="更新间隔（分钟）"
                type="number"
                value={editSubscription.update_interval / 60}
                onChange={(e) => setEditSubscription({
                  ...editSubscription,
                  update_interval: e.target.value * 60,
                })}
                fullWidth
                size="small"
                helperText="设置自动更新的时间间隔"
                disabled={!editSubscription.auto_update}
              />
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>取消</Button>
          <Button variant="contained" onClick={handleSave}>
            保存
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Subscriptions; 