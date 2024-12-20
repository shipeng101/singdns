import React, { useState, useEffect, useCallback } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  TextField,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  IconButton,
  Grid,
  Snackbar,
  Alert,
  CircularProgress,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  Collapse,
  Chip,
  Tooltip,
  LinearProgress,
  Select,
  MenuItem,
  FormControl,
  InputLabel
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Refresh as RefreshIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon
} from '@mui/icons-material';
import * as api from '../services/api';
import usePolling from '../hooks/usePolling';

function Subscriptions() {
  const [subscriptions, setSubscriptions] = useState([]);
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [selectedSubscription, setSelectedSubscription] = useState(null);
  const [formData, setFormData] = useState({ name: '', url: '', groupName: '' });
  const [expandedGroups, setExpandedGroups] = useState({});
  const [updatingNodes, setUpdatingNodes] = useState({});

  // 获取订阅和分组数据
  const fetchData = useCallback(async () => {
    try {
      const [subsResponse, groupsResponse] = await Promise.all([
        api.getSubscriptions(),
        api.getGroups()
      ]);
      setSubscriptions(subsResponse.data);
      setGroups(groupsResponse.data);
    } catch (err) {
      setError(err.response?.data?.message || '获取数据失败');
    }
  }, []);

  // 使用轮询进行实时更新
  usePolling(fetchData, 10000);

  useEffect(() => {
    setLoading(true);
    fetchData().finally(() => setLoading(false));
  }, [fetchData]);

  const showSuccess = (message) => {
    setSuccess(message);
    setTimeout(() => setSuccess(null), 3000);
  };

  const handleAddSubscription = () => {
    setSelectedSubscription(null);
    setFormData({ name: '', url: '', groupName: '' });
    setOpenDialog(true);
  };

  const handleEditSubscription = (subscription) => {
    setSelectedSubscription(subscription);
    setFormData({
      name: subscription.name,
      url: subscription.url,
      groupName: subscription.groupName || ''
    });
    setOpenDialog(true);
  };

  const handleSaveSubscription = async () => {
    if (!formData.name || !formData.url) {
      setError('请填写名称和URL');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      if (selectedSubscription) {
        await api.updateSubscription(selectedSubscription.id, formData);
        showSuccess('订阅已更新');
      } else {
        await api.createSubscription(formData);
        showSuccess('订阅已添加');
      }
      await fetchData();
      setOpenDialog(false);
    } catch (err) {
      setError(err.response?.data?.message || '保存失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteSubscription = async (id) => {
    if (!window.confirm('确定要删除这个订阅吗？')) return;
    setLoading(true);
    setError(null);
    try {
      await api.deleteSubscription(id);
      showSuccess('订阅已删除');
      await fetchData();
    } catch (err) {
      setError(err.response?.data?.message || '删除失败');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateNodes = async (id) => {
    setUpdatingNodes(prev => ({ ...prev, [id]: true }));
    setError(null);
    try {
      await api.updateSubscriptionNodes(id);
      showSuccess('节点已更新');
      await fetchData();
    } catch (err) {
      setError(err.response?.data?.message || '更新节点失败');
    } finally {
      setUpdatingNodes(prev => ({ ...prev, [id]: false }));
    }
  };

  const toggleGroupExpand = (groupId) => {
    setExpandedGroups(prev => ({
      ...prev,
      [groupId]: !prev[groupId]
    }));
  };

  return (
    <Box>
      <Box sx={{ mb: 3, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h5">订阅管理</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleAddSubscription}
          disabled={loading}
        >
          添加订阅
        </Button>
      </Box>

      {loading && <LinearProgress sx={{ mb: 2 }} />}

      <Grid container spacing={2}>
        {subscriptions.map((subscription) => (
          <Grid item xs={12} key={subscription.id}>
            <Card sx={{
              '&:hover': {
                boxShadow: (theme) => theme.shadows[4],
                transform: 'translateY(-2px)',
                transition: 'all 0.3s'
              }
            }}>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
                  <Typography variant="h6">{subscription.name}</Typography>
                  <Box>
                    <Tooltip title="更新节点">
                      <IconButton 
                        onClick={() => handleUpdateNodes(subscription.id)} 
                        disabled={loading || updatingNodes[subscription.id]}
                      >
                        {updatingNodes[subscription.id] ? (
                          <CircularProgress size={24} />
                        ) : (
                          <RefreshIcon />
                        )}
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="编辑">
                      <IconButton onClick={() => handleEditSubscription(subscription)} disabled={loading}>
                        <EditIcon />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="删除">
                      <IconButton onClick={() => handleDeleteSubscription(subscription.id)} disabled={loading}>
                        <DeleteIcon />
                      </IconButton>
                    </Tooltip>
                  </Box>
                </Box>
                <Typography color="textSecondary" sx={{ mb: 1, wordBreak: 'break-all' }}>
                  {subscription.url}
                </Typography>
                {subscription.groupName && (
                  <Chip 
                    label={subscription.groupName} 
                    size="small" 
                    sx={{ mb: 2 }}
                    color="primary"
                    variant="outlined"
                  />
                )}
                <Box sx={{ mt: 2 }}>
                  <Typography variant="subtitle2" sx={{ mb: 1 }}>
                    节点列表 ({subscription.nodes?.length || 0})
                  </Typography>
                  <List dense>
                    {subscription.nodes?.map((node) => (
                      <ListItem key={node.id}>
                        <ListItemText 
                          primary={node.name}
                          secondary={node.type}
                        />
                        {node.status && (
                          <Tooltip title={node.status === 'ok' ? '可用' : '不可用'}>
                            <Box component="span" sx={{ ml: 1 }}>
                              {node.status === 'ok' ? (
                                <CheckCircleIcon color="success" fontSize="small" />
                              ) : (
                                <ErrorIcon color="error" fontSize="small" />
                              )}
                            </Box>
                          </Tooltip>
                        )}
                      </ListItem>
                    ))}
                  </List>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {selectedSubscription ? '编辑订阅' : '添加订阅'}
        </DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="名称"
            fullWidth
            required
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            error={!formData.name}
            helperText={!formData.name && '请输入名称'}
          />
          <TextField
            margin="dense"
            label="URL"
            fullWidth
            required
            value={formData.url}
            onChange={(e) => setFormData({ ...formData, url: e.target.value })}
            error={!formData.url}
            helperText={!formData.url && '请输入URL'}
          />
          <FormControl fullWidth margin="dense">
            <InputLabel>分组</InputLabel>
            <Select
              value={formData.groupName}
              onChange={(e) => setFormData({ ...formData, groupName: e.target.value })}
              label="分组"
            >
              <MenuItem value="">
                <em>无</em>
              </MenuItem>
              {groups.map((group) => (
                <MenuItem key={group.id} value={group.name}>
                  {group.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>取消</Button>
          <Button onClick={handleSaveSubscription} disabled={loading || !formData.name || !formData.url}>
            保存
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={!!error}
        autoHideDuration={6000}
        onClose={() => setError(null)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={() => setError(null)} severity="error">
          {error}
        </Alert>
      </Snackbar>

      <Snackbar
        open={!!success}
        autoHideDuration={3000}
        onClose={() => setSuccess(null)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={() => setSuccess(null)} severity="success">
          {success}
        </Alert>
      </Snackbar>
    </Box>
  );
}

export default Subscriptions; 