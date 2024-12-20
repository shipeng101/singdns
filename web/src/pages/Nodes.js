import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  IconButton,
  Grid,
  TextField,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Snackbar,
  Alert,
  LinearProgress,
  Tooltip,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Chip
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Speed as SpeedIcon,
  CheckCircle as CheckCircleIcon
} from '@mui/icons-material';
import * as api from '../services/api';

function Nodes() {
  const [nodes, setNodes] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [selectedNode, setSelectedNode] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    type: '',
    server: '',
    port: '',
    password: '',
    settings: {}
  });

  const fetchNodeData = async () => {
    setLoading(true);
    try {
      const response = await api.getNodes();
      setNodes(response.data);
    } catch (err) {
      setError(err.response?.data?.message || '获取节点失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNodeData();
  }, []);

  const showSuccess = (message) => {
    setSuccess(message);
    setTimeout(() => setSuccess(null), 3000);
  };

  const handleAddNode = () => {
    setSelectedNode(null);
    setFormData({
      name: '',
      type: '',
      server: '',
      port: '',
      password: '',
      settings: {}
    });
    setOpenDialog(true);
  };

  const handleEditNode = (node) => {
    setSelectedNode(node);
    setFormData({
      name: node.name,
      type: node.type,
      server: node.server,
      port: node.port,
      password: node.password,
      settings: node.settings || {}
    });
    setOpenDialog(true);
  };

  const handleSaveNode = async () => {
    if (!formData.name || !formData.type || !formData.server || !formData.port) {
      setError('请填写所有必填字段');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      if (selectedNode) {
        await api.updateNode(selectedNode.id, formData);
        showSuccess('节点已更新');
      } else {
        await api.createNode(formData);
        showSuccess('节点已添加');
      }
      await fetchNodeData();
      setOpenDialog(false);
    } catch (err) {
      setError(err.response?.data?.message || '保存失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteNode = async (id) => {
    if (!window.confirm('确定要删除这个节点吗？')) return;
    setLoading(true);
    setError(null);
    try {
      await api.deleteNode(id);
      showSuccess('节点已删除');
      await fetchNodeData();
    } catch (err) {
      setError(err.response?.data?.message || '删除失败');
    } finally {
      setLoading(false);
    }
  };

  const handleTestNode = async (id) => {
    setLoading(true);
    setError(null);
    try {
      await api.testNode(id);
      showSuccess('节点测试完成');
      await fetchNodeData();
    } catch (err) {
      setError(err.response?.data?.message || '测试失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box>
      <Box sx={{ mb: 3, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h5">节点管理</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleAddNode}
          disabled={loading}
        >
          添加节点
        </Button>
      </Box>

      {loading && <LinearProgress sx={{ mb: 2 }} />}

      <Grid container spacing={2}>
        {nodes.map((node) => (
          <Grid item xs={12} sm={6} md={4} key={node.id}>
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
                    {node.name}
                  </Typography>
                  <Box>
                    <Tooltip title="测试">
                      <IconButton onClick={() => handleTestNode(node.id)} disabled={loading}>
                        <SpeedIcon />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="编辑">
                      <IconButton onClick={() => handleEditNode(node)} disabled={loading}>
                        <EditIcon />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="删除">
                      <IconButton onClick={() => handleDeleteNode(node.id)} disabled={loading}>
                        <DeleteIcon />
                      </IconButton>
                    </Tooltip>
                  </Box>
                </Box>
                <Typography color="textSecondary" gutterBottom>
                  类型：{node.type}
                </Typography>
                <Typography color="textSecondary" gutterBottom>
                  服务器：{node.server}
                </Typography>
                <Typography color="textSecondary" gutterBottom>
                  端口：{node.port}
                </Typography>
                {node.latency && (
                  <Typography color="textSecondary" gutterBottom>
                    延迟：{node.latency}ms
                  </Typography>
                )}
                {node.status && (
                  <Box sx={{ mt: 1 }}>
                    <Chip
                      icon={node.status === 'ok' ? <CheckCircleIcon /> : null}
                      label={node.status === 'ok' ? '可用' : '不可用'}
                      color={node.status === 'ok' ? 'success' : 'error'}
                      size="small"
                    />
                  </Box>
                )}
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {selectedNode ? '编辑节点' : '添加节点'}
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
          <FormControl fullWidth margin="dense" required>
            <InputLabel>类型</InputLabel>
            <Select
              value={formData.type}
              onChange={(e) => setFormData({ ...formData, type: e.target.value })}
              label="类型"
            >
              <MenuItem value="shadowsocks">Shadowsocks</MenuItem>
              <MenuItem value="vmess">VMess</MenuItem>
              <MenuItem value="trojan">Trojan</MenuItem>
              <MenuItem value="hysteria2">Hysteria2</MenuItem>
              <MenuItem value="tuic">TUIC</MenuItem>
            </Select>
          </FormControl>
          <TextField
            margin="dense"
            label="服务器"
            fullWidth
            required
            value={formData.server}
            onChange={(e) => setFormData({ ...formData, server: e.target.value })}
            error={!formData.server}
            helperText={!formData.server && '请输入服务器地址'}
          />
          <TextField
            margin="dense"
            label="端口"
            type="number"
            fullWidth
            required
            value={formData.port}
            onChange={(e) => setFormData({ ...formData, port: e.target.value })}
            error={!formData.port}
            helperText={!formData.port && '请输入端口'}
          />
          <TextField
            margin="dense"
            label="密码"
            type="password"
            fullWidth
            value={formData.password}
            onChange={(e) => setFormData({ ...formData, password: e.target.value })}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>取消</Button>
          <Button 
            onClick={handleSaveNode} 
            disabled={loading || !formData.name || !formData.type || !formData.server || !formData.port}
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

export default Nodes; 