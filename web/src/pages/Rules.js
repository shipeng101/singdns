import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  IconButton,
  Grid,
  Switch,
  Snackbar,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  LinearProgress,
  Tooltip,
  Divider
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon
} from '@mui/icons-material';
import * as api from '../services/api';

function Rules() {
  const [rules, setRules] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [selectedRule, setSelectedRule] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    pattern: '',
    target: '',
    enabled: true
  });

  const fetchRules = async () => {
    setLoading(true);
    try {
      const response = await api.getRules();
      setRules(response.data);
    } catch (err) {
      setError(err.response?.data?.message || '获取规则失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRules();
  }, []);

  const showSuccess = (message) => {
    setSuccess(message);
    setTimeout(() => setSuccess(null), 3000);
  };

  const handleAddRule = () => {
    setSelectedRule(null);
    setFormData({
      name: '',
      pattern: '',
      target: '',
      enabled: true
    });
    setOpenDialog(true);
  };

  const handleEditRule = (rule) => {
    setSelectedRule(rule);
    setFormData({
      name: rule.name,
      pattern: rule.pattern,
      target: rule.target,
      enabled: rule.enabled
    });
    setOpenDialog(true);
  };

  const handleSaveRule = async () => {
    if (!formData.name || !formData.pattern || !formData.target) {
      setError('请填写所有必填字段');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      if (selectedRule) {
        await api.updateRule(selectedRule.id, formData);
        showSuccess('规则已更新');
      } else {
        await api.createRule(formData);
        showSuccess('规则已添加');
      }
      await fetchRules();
      setOpenDialog(false);
    } catch (err) {
      setError(err.response?.data?.message || '保存失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteRule = async (id) => {
    if (!window.confirm('确定要删除这条规则吗？')) return;
    setLoading(true);
    setError(null);
    try {
      await api.deleteRule(id);
      showSuccess('规则已删除');
      await fetchRules();
    } catch (err) {
      setError(err.response?.data?.message || '删除失败');
    } finally {
      setLoading(false);
    }
  };

  const handleToggleRule = async (id, enabled) => {
    setLoading(true);
    setError(null);
    try {
      await api.updateRule(id, { enabled: !enabled });
      showSuccess(enabled ? '规则已禁用' : '规则已启用');
      await fetchRules();
    } catch (err) {
      setError(err.response?.data?.message || '更新失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box>
      <Box sx={{ mb: 3, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h5">规则管理</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleAddRule}
          disabled={loading}
        >
          添加规则
        </Button>
      </Box>

      {loading && <LinearProgress sx={{ mb: 2 }} />}

      <Grid container spacing={2}>
        {rules.map((rule) => (
          <Grid item xs={12} key={rule.id}>
            <Card sx={{
              '&:hover': {
                boxShadow: (theme) => theme.shadows[4],
                transform: 'translateY(-2px)',
                transition: 'all 0.3s'
              }
            }}>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h6" component="div">
                    {rule.name}
                  </Typography>
                  <Box>
                    <Tooltip title={rule.enabled ? '禁用' : '启用'}>
                      <Switch
                        checked={rule.enabled}
                        onChange={() => handleToggleRule(rule.id, rule.enabled)}
                        disabled={loading}
                      />
                    </Tooltip>
                    <Tooltip title="编辑">
                      <IconButton onClick={() => handleEditRule(rule)} disabled={loading}>
                        <EditIcon />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="删除">
                      <IconButton onClick={() => handleDeleteRule(rule.id)} disabled={loading}>
                        <DeleteIcon />
                      </IconButton>
                    </Tooltip>
                  </Box>
                </Box>
                <Divider sx={{ my: 1 }} />
                <Typography color="textSecondary" sx={{ mt: 1 }}>
                  匹配模式：{rule.pattern}
                </Typography>
                <Typography color="textSecondary" sx={{ mt: 1 }}>
                  目标节点：{rule.target}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {selectedRule ? '编辑规则' : '添加规则'}
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
            label="匹配模式"
            fullWidth
            required
            value={formData.pattern}
            onChange={(e) => setFormData({ ...formData, pattern: e.target.value })}
            error={!formData.pattern}
            helperText={!formData.pattern && '请输入匹配模式'}
          />
          <TextField
            margin="dense"
            label="目标节点"
            fullWidth
            required
            value={formData.target}
            onChange={(e) => setFormData({ ...formData, target: e.target.value })}
            error={!formData.target}
            helperText={!formData.target && '请输入目标节点'}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>取消</Button>
          <Button 
            onClick={handleSaveRule} 
            disabled={loading || !formData.name || !formData.pattern || !formData.target}
          >
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

export default Rules; 