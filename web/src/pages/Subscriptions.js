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
  Divider
} from '@mui/material';
import {
  DarkMode as DarkModeIcon,
  LightMode as LightModeIcon,
  OpenInNew as OpenInNewIcon,
  Add as AddIcon,
  Delete as DeleteIcon,
  Edit as EditIcon,
  Sync as SyncIcon
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
      setSubscriptions(data);
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
      setLoading(true);
      const subscriptionData = {
        ...editSubscription,
        updateInterval: editSubscription.autoUpdate ? formatUpdateInterval(editSubscription.updateInterval) : 0
      };

      if (selectedSubscription) {
        await updateSubscription(selectedSubscription.id, subscriptionData);
      } else {
        await createSubscription(subscriptionData);
      }
      await fetchData();
      setSuccess(true);
      handleCloseDialog();
    } catch (err) {
      setError(err.response?.data?.error || '保存订阅失败');
    } finally {
      setLoading(false);
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

      <Card sx={{ ...styles.card, mt: 1.5 }}>
        <CardContent>
          <Stack direction="row" justifyContent="space-between" alignItems="center" mb={2}>
            <Typography variant="h6">订阅列表</Typography>
            <Button
              variant="contained"
              size="small"
              startIcon={<AddIcon />}
              onClick={() => handleOpenDialog()}
              sx={styles.actionButton}
            >
              添加订阅
            </Button>
          </Stack>

          <List sx={{ width: '100%' }}>
            {subscriptions.map((subscription, index) => (
              <React.Fragment key={subscription.id}>
                {index > 0 && <Divider component="li" />}
                <ListItem
                  sx={{
                    py: 2,
                    px: 0,
                    '&:hover': {
                      backgroundColor: 'action.hover',
                      borderRadius: 1
                    }
                  }}
                >
                  <ListItemIcon>
                    <Chip 
                      label={subscription.type} 
                      size="small"
                      sx={{
                        ...styles.chip.primary,
                        background: 'linear-gradient(45deg, #3949ab 30%, #5e35b1 90%)',
                        color: '#fff',
                        fontWeight: 500,
                      }}
                    />
                  </ListItemIcon>
                  <ListItemText
                    primary={
                      <Stack direction="row" spacing={1} alignItems="center">
                        <Typography variant="subtitle1" sx={{ fontWeight: 500 }}>
                          {subscription.name}
                        </Typography>
                        <Chip 
                          label={subscription.active ? "已启用" : "已禁用"} 
                          size="small"
                          sx={subscription.active ? {
                            ...styles.chip.success,
                            background: 'linear-gradient(45deg, #4caf50 30%, #81c784 90%)',
                            color: '#fff',
                            fontWeight: 500,
                          } : {
                            ...styles.chip.error,
                            background: 'linear-gradient(45deg, #f44336 30%, #e57373 90%)',
                            color: '#fff',
                            fontWeight: 500,
                          }}
                        />
                        {subscription.autoUpdate && (
                          <Chip 
                            label={`自动更新: ${subscription.updateInterval}分钟`} 
                            size="small"
                            sx={{
                              ...styles.chip.warning,
                              background: 'linear-gradient(45deg, #ff9800 30%, #ffb74d 90%)',
                              color: '#fff',
                              fontWeight: 500,
                            }}
                          />
                        )}
                      </Stack>
                    }
                    secondary={subscription.url}
                    secondaryTypographyProps={{
                      sx: {
                        mt: 0.5,
                        color: 'text.secondary',
                        wordBreak: 'break-all'
                      }
                    }}
                  />
                  <ListItemSecondaryAction>
                    <Stack direction="row" spacing={1}>
                      <IconButton
                        edge="end"
                        size="small"
                        onClick={() => handleUpdate(subscription.id)}
                        sx={{
                          color: 'text.secondary',
                          '&:hover': {
                            color: 'primary.main',
                            backgroundColor: 'action.hover',
                          },
                        }}
                      >
                        <SyncIcon fontSize="small" />
                      </IconButton>
                      <IconButton
                        edge="end"
                        size="small"
                        onClick={() => handleOpenDialog(subscription)}
                        sx={{
                          color: 'text.secondary',
                          '&:hover': {
                            color: 'primary.main',
                            backgroundColor: 'action.hover',
                          },
                        }}
                      >
                        <EditIcon fontSize="small" />
                      </IconButton>
                      <IconButton
                        edge="end"
                        size="small"
                        onClick={() => handleDelete(subscription.id)}
                        sx={{
                          color: 'text.secondary',
                          '&:hover': {
                            color: 'error.main',
                            backgroundColor: 'action.hover',
                          },
                        }}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Stack>
                  </ListItemSecondaryAction>
                </ListItem>
              </React.Fragment>
            ))}
          </List>
        </CardContent>
      </Card>

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
            <FormControl fullWidth size="small">
              <InputLabel>订阅类型</InputLabel>
              <Select
                value={editSubscription.type}
                label="订阅类型"
                onChange={(e) => setEditSubscription({ ...editSubscription, type: e.target.value })}
              >
                {subscriptionTypes.map((type) => (
                  <MenuItem key={type.value} value={type.value}>
                    <Stack>
                      <Typography variant="body1">{type.label}</Typography>
                      <Typography variant="caption" color="text.secondary">
                        {type.description}
                      </Typography>
                    </Stack>
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <TextField
              fullWidth
              label="订阅地址"
              value={editSubscription.url}
              onChange={(e) => setEditSubscription({ ...editSubscription, url: e.target.value })}
              size="small"
              placeholder="https://example.com/subscribe"
            />
            <FormControlLabel
              control={
                <Switch
                  checked={editSubscription.autoUpdate}
                  onChange={(e) => setEditSubscription({ ...editSubscription, autoUpdate: e.target.checked })}
                />
              }
              label="自动更新"
            />
            {editSubscription.autoUpdate && (
              <Stack direction="row" spacing={1}>
                <TextField
                  type="number"
                  label="更新间隔"
                  value={editSubscription.updateInterval.value}
                  onChange={(e) => setEditSubscription({
                    ...editSubscription,
                    updateInterval: {
                      ...editSubscription.updateInterval,
                      value: parseInt(e.target.value) || 1
                    }
                  })}
                  size="small"
                  sx={{ width: '40%' }}
                  InputProps={{
                    inputProps: { min: 1 }
                  }}
                />
                <FormControl size="small" sx={{ width: '60%' }}>
                  <InputLabel>时间单位</InputLabel>
                  <Select
                    value={editSubscription.updateInterval.unit}
                    label="时间单位"
                    onChange={(e) => setEditSubscription({
                      ...editSubscription,
                      updateInterval: {
                        ...editSubscription.updateInterval,
                        unit: e.target.value
                      }
                    })}
                  >
                    {updateIntervalUnits.map((unit) => (
                      <MenuItem key={unit.value} value={unit.value}>{unit.label}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Stack>
            )}
            <FormControlLabel
              control={
                <Switch
                  checked={editSubscription.active}
                  onChange={(e) => setEditSubscription({ ...editSubscription, active: e.target.checked })}
                />
              }
              label="启用订阅"
            />
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