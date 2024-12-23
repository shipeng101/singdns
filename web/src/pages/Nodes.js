import React, { useState, useEffect, useCallback } from 'react';
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
  Chip,
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
  FormHelperText,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Divider,
  Pagination,
  Checkbox,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  DarkMode as DarkModeIcon,
  LightMode as LightModeIcon,
  OpenInNew as OpenInNewIcon,
  Sync as SyncIcon,
} from '@mui/icons-material';
import { motion } from 'framer-motion';
import { getNodes, getNodeGroups, createNodeGroup, updateNodeGroup, deleteNodeGroup, createNode, updateNode, deleteNode, testNode, testNodes } from '../services/api';
import { getCommonStyles } from '../styles/commonStyles';

const Nodes = ({ mode, setMode, onDashboardClick }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);
  const [nodes, setNodes] = useState([]);
  const [nodeGroups, setNodeGroups] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [openGroupDialog, setOpenGroupDialog] = useState(false);
  const [selectedGroup, setSelectedGroup] = useState(null);
  const [editGroup, setEditGroup] = useState({
    name: '',
    tag: '',
    includePatterns: [],
    excludePatterns: [],
    mode: 'select',
    active: true
  });
  const [openNodeDialog, setOpenNodeDialog] = useState(false);
  const [selectedNode, setSelectedNode] = useState(null);
  const [editNode, setEditNode] = useState({
    name: '',
    type: 'ss',
    address: '',
    port: '',
    method: 'aes-256-gcm',
    password: '',
    uuid: '',
    alterId: 0,
    security: 'auto',
    host: '',
    skipCertVerify: false,
    protocol: 'udp',
    up: '',
    down: '',
    obfs: '',
    alpn: [],
    upMbps: 100,
    downMbps: 100,
    version: 3,
    username: '',
    tls: false,
    network: 'tcp',
    udp: true,
    webSocket: false,
    path: '',
    wsHeaders: {},
    plugin: '',
    pluginOpts: {},
    cc: '',
    serviceName: ''
  });
  const [selectedNodes, setSelectedNodes] = useState([]);

  const nodeTypes = [
    { value: 'ss', label: 'Shadowsocks' },
    { value: 'vmess', label: 'VMess' },
    { value: 'trojan', label: 'Trojan' },
    { value: 'naive', label: 'Naive' },
    { value: 'hysteria', label: 'Hysteria' },
    { value: 'hysteria2', label: 'Hysteria2' },
    { value: 'shadowtls', label: 'ShadowTLS' },
    { value: 'tun', label: 'Tun' },
    { value: 'redirect', label: 'Redirect' },
    { value: 'tproxy', label: 'TProxy' },
    { value: 'socks', label: 'Socks' },
    { value: 'http', label: 'HTTP' },
  ];

  const ssEncryption = [
    'aes-128-gcm',
    'aes-256-gcm',
    'chacha20-poly1305',
    'xchacha20-poly1305',
    '2022-blake3-aes-128-gcm',
    '2022-blake3-aes-256-gcm',
    '2022-blake3-chacha20-poly1305',
  ];

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [nodesResponse, groupsData] = await Promise.all([
        getNodes(page, pageSize),
        getNodeGroups()
      ]);

      // Sort nodes by latency (offline nodes at the end)
      const sortedNodes = nodesResponse.nodes.sort((a, b) => {
        if (a.status === 'offline' && b.status === 'offline') return 0;
        if (a.status === 'offline') return 1;
        if (b.status === 'offline') return -1;
        return a.latency - b.latency;
      });

      setNodes(sortedNodes);
      setTotal(nodesResponse.pagination.total);
      setNodeGroups(groupsData);
      setError(null);
    } catch (err) {
      setError(err.response?.data?.error || '获取数据失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleRefresh = async () => {
    try {
      setLoading(true);
      // 测试所有节点
      const response = await testNodes();
      // 更新节点状态
      const updatedNodes = nodes.map(node => {
        const result = response.results.find(r => r.id === node.id);
        if (result) {
          return {
            ...node,
            status: result.status,
            latency: result.latency,
            checkedAt: result.checkedAt || new Date().toISOString()
          };
        }
        return node;
      });

      // 按延迟排序，离线节点放在最后
      const sortedNodes = updatedNodes.sort((a, b) => {
        if (a.status === 'offline' && b.status === 'offline') return 0;
        if (a.status === 'offline') return 1;
        if (b.status === 'offline') return -1;
        return a.latency - b.latency;
      });

      setNodes(sortedNodes);
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '刷新节点失败');
    } finally {
      setLoading(false);
    }
  };

  const handlePageChange = (event, newPage) => {
    setPage(newPage);
  };

  const handleOpenGroupDialog = (group = null) => {
    if (group) {
      setSelectedGroup(group);
      setEditGroup({
        ...group,
        includePatterns: group.includePatterns || [],
        excludePatterns: group.excludePatterns || [],
        mode: group.mode || 'select',
        tag: group.tag || '',
        active: group.active ?? true
      });
    } else {
      setSelectedGroup(null);
      setEditGroup({
        name: '',
        tag: '',
        includePatterns: [],
        excludePatterns: [],
        mode: 'select',
        active: true
      });
    }
    setOpenGroupDialog(true);
  };

  const handleCloseGroupDialog = () => {
    setOpenGroupDialog(false);
    setSelectedGroup(null);
    setEditGroup({
      name: '',
      tag: '',
      includePatterns: [],
      excludePatterns: [],
      mode: 'select',
      active: true
    });
  };

  const handleSaveGroup = async () => {
    try {
      setLoading(true);
      if (selectedGroup) {
        await updateNodeGroup(selectedGroup.id, editGroup);
      } else {
        await createNodeGroup(editGroup);
      }
      await fetchData();
      setSuccess(true);
      handleCloseGroupDialog();
    } catch (err) {
      setError(err.response?.data?.error || '保存分组失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteGroup = async (id) => {
    try {
      setLoading(true);
      await deleteNodeGroup(id);
      await fetchData();
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '删除分组失败');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenNodeDialog = (node = null) => {
    if (node) {
      setSelectedNode(node);
      setEditNode({
        ...node,
        webSocket: node.network === 'ws',
      });
    } else {
      setSelectedNode(null);
      setEditNode({
        name: '',
        type: 'ss',
        address: '',
        port: '',
        method: 'aes-256-gcm',
        password: '',
        uuid: '',
        alterId: 0,
        security: 'auto',
        host: '',
        skipCertVerify: false,
        protocol: 'udp',
        up: '',
        down: '',
        obfs: '',
        alpn: [],
        upMbps: 100,
        downMbps: 100,
        version: 3,
        username: '',
        tls: false,
        network: 'tcp',
        udp: true,
        webSocket: false,
        path: '',
        wsHeaders: {},
        plugin: '',
        pluginOpts: {},
        cc: '',
        serviceName: ''
      });
    }
    setOpenNodeDialog(true);
  };

  const handleCloseNodeDialog = () => {
    setOpenNodeDialog(false);
    setSelectedNode(null);
    setEditNode({
      name: '',
      type: 'ss',
      address: '',
      port: '',
      method: 'aes-256-gcm',
      password: '',
      uuid: '',
      alterId: 0,
      security: 'auto',
      host: '',
      skipCertVerify: false,
      protocol: 'udp',
      up: '',
      down: '',
      obfs: '',
      alpn: [],
      upMbps: 100,
      downMbps: 100,
      version: 3,
      username: '',
      tls: false,
      network: 'tcp',
      udp: true,
      webSocket: false,
      path: '',
      wsHeaders: {},
      plugin: '',
      pluginOpts: {},
      cc: '',
      serviceName: ''
    });
  };

  const handleSaveNode = async () => {
    try {
      setLoading(true);
      const nodeData = {
        ...editNode,
        id: selectedNode ? selectedNode.id : undefined,
        group: selectedNode ? selectedNode.group : '',
        status: selectedNode ? selectedNode.status : 'offline',
        latency: selectedNode ? selectedNode.latency : 0,
        checkedAt: selectedNode ? selectedNode.checkedAt : new Date().toISOString(),
        createdAt: selectedNode ? selectedNode.createdAt : Math.floor(Date.now() / 1000),
        updatedAt: Math.floor(Date.now() / 1000)
      };

      if (selectedNode) {
        await updateNode(selectedNode.id, nodeData);
      } else {
        await createNode(nodeData);
      }
      await fetchData();
      setSuccess(true);
      handleCloseNodeDialog();
    } catch (err) {
      setError(err.response?.data?.error || '保存节点失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteNode = async (id) => {
    if (!window.confirm('确定要删除这个节点吗？')) {
      return;
    }
    
    try {
      setLoading(true);
      await deleteNode(id);
      await fetchData();
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '删除节点失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSelectAll = (event) => {
    if (event.target.checked) {
      setSelectedNodes(nodes.map(node => node.id));
    } else {
      setSelectedNodes([]);
    }
  };

  const handleSelectNode = (nodeId) => {
    setSelectedNodes(prev => {
      if (prev.includes(nodeId)) {
        return prev.filter(id => id !== nodeId);
      } else {
        return [...prev, nodeId];
      }
    });
  };

  const handleBatchDelete = async () => {
    if (selectedNodes.length === 0) {
      setError('请先选择要删除的节点');
      return;
    }

    if (!window.confirm(`确定要删除选中的 ${selectedNodes.length} 个节点吗？`)) {
      return;
    }

    try {
      setLoading(true);
      // 依次删除选中的节点
      for (const nodeId of selectedNodes) {
        await deleteNode(nodeId);
      }
      // 刷新节点列表
      await fetchData();
      setSelectedNodes([]); // 清空选中状态
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '批量删除节点失败');
    } finally {
      setLoading(false);
    }
  };

  const renderNodeFields = () => {
    switch (editNode.type) {
      case 'ss':
        return (
          <>
            <Grid item xs={12}>
              <FormControl fullWidth size="small">
                <InputLabel>加密方式</InputLabel>
                <Select
                  value={editNode.method}
                  label="加密方式"
                  onChange={(e) => setEditNode({ ...editNode, method: e.target.value })}
                  sx={styles.textField}
                >
                  {ssEncryption.map((method) => (
                    <MenuItem key={method} value={method}>{method}</MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="密码"
                value={editNode.password}
                onChange={(e) => setEditNode({ ...editNode, password: e.target.value })}
                size="small"
                type="password"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="插件"
                value={editNode.plugin}
                onChange={(e) => setEditNode({ ...editNode, plugin: e.target.value })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="插件选项"
                value={editNode.pluginOpts}
                onChange={(e) => setEditNode({ ...editNode, pluginOpts: e.target.value })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
          </>
        );
      case 'vmess':
        return (
          <>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="UUID"
                value={editNode.uuid}
                onChange={(e) => setEditNode({ ...editNode, uuid: e.target.value })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="AlterID"
                type="number"
                value={editNode.alterId}
                onChange={(e) => setEditNode({ ...editNode, alterId: parseInt(e.target.value) || 0 })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12}>
              <FormControl fullWidth size="small">
                <InputLabel>加密方式</InputLabel>
                <Select
                  value={editNode.security}
                  label="加密方式"
                  onChange={(e) => setEditNode({ ...editNode, security: e.target.value })}
                  sx={styles.textField}
                >
                  <MenuItem value="auto">auto</MenuItem>
                  <MenuItem value="aes-128-gcm">aes-128-gcm</MenuItem>
                  <MenuItem value="chacha20-poly1305">chacha20-poly1305</MenuItem>
                  <MenuItem value="none">none</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12}>
              <FormControl fullWidth size="small">
                <InputLabel>传输协议</InputLabel>
                <Select
                  value={editNode.network}
                  label="传输协议"
                  onChange={(e) => setEditNode({ ...editNode, network: e.target.value })}
                  sx={styles.textField}
                >
                  <MenuItem value="tcp">TCP</MenuItem>
                  <MenuItem value="ws">WebSocket</MenuItem>
                  <MenuItem value="http">HTTP</MenuItem>
                  <MenuItem value="h2">HTTP/2</MenuItem>
                  <MenuItem value="grpc">gRPC</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            {editNode.network === 'ws' && (
              <>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="WebSocket 路径"
                    value={editNode.path}
                    onChange={(e) => setEditNode({ ...editNode, path: e.target.value })}
                    size="small"
                    sx={styles.textField}
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="WebSocket 主机"
                    value={editNode.host}
                    onChange={(e) => setEditNode({ ...editNode, host: e.target.value })}
                    size="small"
                    sx={styles.textField}
                  />
                </Grid>
              </>
            )}
          </>
        );
      case 'trojan':
        return (
          <>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="密码"
                value={editNode.password}
                onChange={(e) => setEditNode({ ...editNode, password: e.target.value })}
                size="small"
                type="password"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="SNI"
                value={editNode.host}
                onChange={(e) => setEditNode({ ...editNode, host: e.target.value })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
          </>
        );
      default:
        return null;
    }
  };

  // 添加 URL 解码函数
  const decodeNodeName = (name) => {
    try {
      return decodeURIComponent(name);
    } catch (e) {
      return name;
    }
  };

  // 格式化延时显示
  const formatLatency = (latency) => {
    if (latency === undefined || latency === null) return '-';
    if (latency === 0) return '-';
    if (latency < 0) return '超时';
    return `${latency}ms`;
  };

  // 获取节点状态的颜色
  const getStatusColor = (status, latency) => {
    if (status === 'online') {
      if (latency < 100) {
        return 'linear-gradient(45deg, #4caf50 30%, #81c784 90%)';  // 绿色
      } else if (latency < 300) {
        return 'linear-gradient(45deg, #ff9800 30%, #ffb74d 90%)';  // 橙色
      } else {
        return 'linear-gradient(45deg, #f44336 30%, #e57373 90%)';  // 红色
      }
    }
    return 'linear-gradient(45deg, #9e9e9e 30%, #bdbdbd 90%)';  // 灰色
  };

  const columns = [
    { field: 'name', headerName: '名称', flex: 1 },
    { field: 'type', headerName: '类型', width: 100 },
    { field: 'address', headerName: '地址', flex: 1 },
    { field: 'port', headerName: '端口', width: 100 },
    {
      field: 'status',
      headerName: '状态',
      width: 120,
      renderCell: (params) => (
        <Chip
          label={params.value === 'online' ? '在线' : '离线'}
          color={params.value === 'online' ? 'success' : 'error'}
          size="small"
          sx={{
            background: getStatusColor(params.value, params.row.latency),
            color: 'white'
          }}
        />
      ),
    },
    {
      field: 'latency',
      headerName: '延迟',
      width: 100,
      renderCell: (params) => (
        <span>{formatLatency(params.value)}</span>
      ),
    },
    {
      field: 'checkedAt',
      headerName: '最后测速',
      width: 180,
      renderCell: (params) => (
        params.value ? new Date(params.value).toLocaleString('zh-CN', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
          hour12: false
        }) : '-'
      ),
    },
    {
      field: 'actions',
      headerName: '操作',
      width: 120,
      renderCell: (params) => (
        <Stack direction="row" spacing={1}>
          <IconButton
            size="small"
            onClick={() => handleOpenNodeDialog(params.row)}
          >
            <EditIcon fontSize="small" />
          </IconButton>
          <IconButton
            size="small"
            onClick={() => handleDeleteNode(params.row.id)}
          >
            <DeleteIcon fontSize="small" />
          </IconButton>
        </Stack>
      ),
    },
  ];

  // 修复中文编码问题
  const renderGroupModeHelperText = (mode) => {
    return mode === 'select' ? '手动选择节点' : '自动选择可用和延迟较低的节点';
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
                节点管理
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

      {/* Nodes List */}
      <Box sx={{ mb: 3 }}>
        <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Typography variant="h6">节点列表</Typography>
          <Stack direction="row" spacing={1}>
            <Button
              variant="contained"
              size="small"
              startIcon={<AddIcon />}
              onClick={() => handleOpenNodeDialog()}
              sx={styles.actionButton}
            >
              添加节点
            </Button>
            <Button
              variant="contained"
              size="small"
              startIcon={<SyncIcon />}
              onClick={handleRefresh}
              disabled={loading}
              sx={styles.actionButton}
            >
              刷新节点
            </Button>
            {selectedNodes.length > 0 && (
              <Button
                variant="contained"
                size="small"
                color="error"
                startIcon={<DeleteIcon />}
                onClick={handleBatchDelete}
                disabled={loading}
                sx={styles.actionButton}
              >
                删除选中 ({selectedNodes.length})
              </Button>
            )}
          </Stack>
        </Box>

        <TableContainer component={Paper} sx={styles.tableContainer}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell padding="checkbox">
                  <Checkbox
                    indeterminate={selectedNodes.length > 0 && selectedNodes.length < nodes.length}
                    checked={nodes.length > 0 && selectedNodes.length === nodes.length}
                    onChange={handleSelectAll}
                    size="small"
                  />
                </TableCell>
                <TableCell>名称</TableCell>
                <TableCell>类型</TableCell>
                <TableCell>地址</TableCell>
                <TableCell>端口</TableCell>
                <TableCell>延迟</TableCell>
                <TableCell>状态</TableCell>
                <TableCell>最后测速</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {nodes.map((node) => (
                <TableRow key={node.id}>
                  <TableCell padding="checkbox">
                    <Checkbox
                      checked={selectedNodes.includes(node.id)}
                      onChange={() => handleSelectNode(node.id)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{decodeNodeName(node.name)}</TableCell>
                  <TableCell>{node.type}</TableCell>
                  <TableCell>{node.address}</TableCell>
                  <TableCell>{node.port}</TableCell>
                  <TableCell>
                    <Chip
                      label={formatLatency(node.latency)}
                      size="small"
                      sx={{
                        background: getStatusColor(node.status, node.latency),
                        color: '#fff',
                        fontWeight: 500,
                      }}
                    />
                  </TableCell>
                  <TableCell>
                    <Chip 
                      label={node.status === 'online' ? '在线' : '离线'} 
                      size="small"
                      sx={{
                        background: node.status === 'online' 
                          ? 'linear-gradient(45deg, #4caf50 30%, #81c784 90%)'
                          : 'linear-gradient(45deg, #f44336 30%, #e57373 90%)',
                        color: '#fff',
                        fontWeight: 500,
                      }}
                    />
                  </TableCell>
                  <TableCell>
                    {node.checkedAt ? new Date(node.checkedAt).toLocaleString('zh-CN', {
                      year: 'numeric',
                      month: '2-digit',
                      day: '2-digit',
                      hour: '2-digit',
                      minute: '2-digit',
                      second: '2-digit',
                      hour12: false
                    }) : '-'}
                  </TableCell>
                  <TableCell align="right">
                    <Stack direction="row" spacing={0.5} justifyContent="flex-end">
                      <IconButton 
                        size="small" 
                        onClick={() => handleOpenNodeDialog(node)}
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
                        size="small" 
                        onClick={() => handleDeleteNode(node.id)}
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
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>

        <Box sx={{ mt: 2, display: 'flex', justifyContent: 'center' }}>
          <Pagination
            count={Math.ceil(total / pageSize)}
            page={page}
            onChange={handlePageChange}
            color="primary"
            size="small"
          />
        </Box>
      </Box>

      {/* Node Groups */}
      <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h6">节点分组</Typography>
        <Button
          variant="contained"
          size="small"
          startIcon={<AddIcon />}
          onClick={() => handleOpenGroupDialog()}
          sx={styles.actionButton}
        >
          添加分组
        </Button>
      </Box>

      <Grid container spacing={1.5}>
        {nodeGroups.map((group) => (
          <Grid item xs={12} sm={6} md={3} key={group.id}>
            <Card sx={styles.card}>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1.5 }}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Chip 
                      label={group.tag} 
                      size="small"
                      sx={{
                        background: 'linear-gradient(45deg, #3949ab 30%, #5e35b1 90%)',
                        color: '#fff',
                        fontWeight: 500,
                      }}
                    />
                    <Typography variant="h6">{group.name}</Typography>
                  </Stack>
                  <Stack direction="row" spacing={0.5}>
                    <IconButton 
                      size="small" 
                      onClick={() => handleOpenGroupDialog(group)}
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
                      size="small" 
                      onClick={() => handleDeleteGroup(group.id)}
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
                </Box>
                <Stack direction="row" spacing={1} alignItems="center">
                  <Typography variant="body2" color="text.secondary">
                    {group.nodeCount || 0} 个节点
                  </Typography>
                  <Chip 
                    label={group.active ? "活跃" : "未激活"} 
                    size="small"
                    sx={group.active ? {
                      background: 'linear-gradient(45deg, #4caf50 30%, #81c784 90%)',
                      color: '#fff',
                      fontWeight: 500,
                    } : {
                      background: 'linear-gradient(45deg, #f44336 30%, #e57373 90%)',
                      color: '#fff',
                      fontWeight: 500,
                    }}
                  />
                  <Chip 
                    label={group.mode === 'select' ? "手动选择" : "自动测速"} 
                    size="small"
                    sx={group.mode === 'select' ? {
                      background: 'linear-gradient(45deg, #4caf50 30%, #81c784 90%)',
                      color: '#fff',
                      fontWeight: 500,
                    } : {
                      background: 'linear-gradient(45deg, #ff9800 30%, #ffb74d 90%)',
                      color: '#fff',
                      fontWeight: 500,
                    }}
                  />
                </Stack>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* Group Dialog */}
      <Dialog 
        open={openGroupDialog} 
        onClose={handleCloseGroupDialog} 
        maxWidth="sm" 
        fullWidth
        PaperProps={{
          sx: styles.dialog
        }}
      >
        <DialogTitle>
          {selectedGroup ? '编辑分组' : '添加分组'}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="分组名称"
                  value={editGroup.name}
                  onChange={(e) => setEditGroup({ ...editGroup, name: e.target.value })}
                  size="small"
                  sx={styles.textField}
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="分组标签"
                  value={editGroup.tag}
                  onChange={(e) => setEditGroup({ ...editGroup, tag: e.target.value })}
                  size="small"
                  placeholder="例如：HK, JP, SG, US"
                  sx={styles.textField}
                />
              </Grid>
              <Grid item xs={12}>
                <FormControl fullWidth size="small">
                  <InputLabel>节点选择模式</InputLabel>
                  <Select
                    value={editGroup.mode}
                    label="节点选择模式"
                    onChange={(e) => setEditGroup({ ...editGroup, mode: e.target.value })}
                    sx={styles.textField}
                  >
                    <MenuItem value="select">手动选择</MenuItem>
                    <MenuItem value="urltest">自动测速</MenuItem>
                  </Select>
                  <FormHelperText>
                    {renderGroupModeHelperText(editGroup.mode)}
                  </FormHelperText>
                </FormControl>
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={editGroup.active}
                      onChange={(e) => setEditGroup({ ...editGroup, active: e.target.checked })}
                    />
                  }
                  label="活跃"
                />
              </Grid>
              <Grid item xs={12}>
                <Typography variant="subtitle2" gutterBottom>
                  节点匹配规则
                </Typography>
                <Stack spacing={2}>
                  <Box>
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      包含节点（白名单模式）
                    </Typography>
                    <TextField
                      fullWidth
                      multiline
                      rows={3}
                      placeholder="每行一个正则表达式，例如：
香港|HK|🇭🇰
新加坡|SG|🇸🇬
日本|JP|🇯🇵
美国|US|🇺🇸"
                      value={editGroup.includePatterns.join('\n')}
                      onChange={(e) => setEditGroup({
                        ...editGroup,
                        includePatterns: e.target.value.split('\n').filter(p => p.trim())
                      })}
                      size="small"
                      sx={styles.textField}
                    />
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      排除节点（黑名单模式）
                    </Typography>
                    <TextField
                      fullWidth
                      multiline
                      rows={3}
                      placeholder="每行一个正则表达式，例如：.*过期.*"
                      value={editGroup.excludePatterns.join('\n')}
                      onChange={(e) => setEditGroup({
                        ...editGroup,
                        excludePatterns: e.target.value.split('\n').filter(p => p.trim())
                      })}
                      size="small"
                      sx={styles.textField}
                    />
                  </Box>
                </Stack>
              </Grid>
            </Grid>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button 
            onClick={handleCloseGroupDialog}
            size="small"
            sx={{
              color: 'text.secondary',
              '&:hover': {
                backgroundColor: 'action.hover',
              },
            }}
          >
            取消
          </Button>
          <Button 
            variant="contained" 
            onClick={handleSaveGroup}
            disabled={loading}
            size="small"
            sx={styles.actionButton}
          >
            保存
          </Button>
        </DialogActions>
      </Dialog>

      {/* Node Dialog */}
      <Dialog
        open={openNodeDialog}
        onClose={handleCloseNodeDialog}
        maxWidth="sm"
        fullWidth
        PaperProps={{
          sx: styles.dialog
        }}
      >
        <DialogTitle>
          {selectedNode ? '编辑节点' : '添加节点'}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="名称"
                  value={editNode.name}
                  onChange={(e) => setEditNode({ ...editNode, name: e.target.value })}
                  size="small"
                  sx={styles.textField}
                />
              </Grid>
              <Grid item xs={12}>
                <FormControl fullWidth size="small">
                  <InputLabel>类型</InputLabel>
                  <Select
                    value={editNode.type}
                    label="类型"
                    onChange={(e) => setEditNode({ ...editNode, type: e.target.value })}
                    sx={styles.textField}
                  >
                    {nodeTypes.map((type) => (
                      <MenuItem key={type.value} value={type.value}>{type.label}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="地址"
                  value={editNode.address}
                  onChange={(e) => setEditNode({ ...editNode, address: e.target.value })}
                  size="small"
                  sx={styles.textField}
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="端口"
                  type="number"
                  value={editNode.port}
                  onChange={(e) => setEditNode({ ...editNode, port: parseInt(e.target.value) || '' })}
                  size="small"
                  sx={styles.textField}
                />
              </Grid>
              {renderNodeFields()}
              <Grid item xs={12}>
                <Divider sx={{ my: 1 }}>高级设置</Divider>
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={editNode.tls}
                      onChange={(e) => setEditNode({ ...editNode, tls: e.target.checked })}
                    />
                  }
                  label="启用 TLS"
                />
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={editNode.skipCertVerify}
                      onChange={(e) => setEditNode({ ...editNode, skipCertVerify: e.target.checked })}
                    />
                  }
                  label="跳过证书验证"
                />
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={editNode.udp}
                      onChange={(e) => setEditNode({ ...editNode, udp: e.target.checked })}
                    />
                  }
                  label="启用 UDP"
                />
              </Grid>
            </Grid>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button 
            onClick={handleCloseNodeDialog}
            size="small"
            sx={{
              color: 'text.secondary',
              '&:hover': {
                backgroundColor: 'action.hover',
              },
            }}
          >
            取消
          </Button>
          <Button 
            variant="contained" 
            onClick={handleSaveNode}
            disabled={loading}
            size="small"
            sx={styles.actionButton}
          >
            保存
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Nodes; 