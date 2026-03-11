// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import React, { useState, useCallback, useEffect } from 'react';
import {
  Box,
  Container,
  Heading,
  Text,
  VStack,
  HStack,
  Link,
  Button,
  Spinner,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  CloseButton,
  SimpleGrid,
  Card,
  CardBody,
  Badge,
  IconButton,
  Tooltip,
  useToast,
} from '@chakra-ui/react';
import { useDropzone } from 'react-dropzone';
import { motion, AnimatePresence } from 'framer-motion';
import ReadingScreen from './components/ReadingScreen.jsx';
import AdSense from './components/AdSense.jsx';
import { parseDocument, isSupported } from './parsers/index.js';
import {
  getRecentSessions,
  saveSession,
  deleteSession,
  getSettings,
  createOrUpdateSession,
} from './storage.js';

const MotionBox = motion(Box);

// Get AdSense config from URL params or environment
function getAdsenseConfig() {
  const params = new URLSearchParams(window.location.search);
  const key = params.get('adsense') || import.meta.env.VITE_ADSENSE_KEY || null;
  return {
    enabled: !!key,
    key,
  };
}

function App() {
  const [view, setView] = useState('home'); // 'home', 'loading', 'reading', 'error'
  const [document, setDocument] = useState(null);
  const [words, setWords] = useState([]);
  const [documentHash, setDocumentHash] = useState(null);
  const [error, setError] = useState(null);
  const [recentSessions, setRecentSessions] = useState([]);
  const [adsenseConfig] = useState(getAdsenseConfig);
  const toast = useToast();

  // Load recent sessions on mount
  useEffect(() => {
    setRecentSessions(getRecentSessions(5));
  }, [view]);

  const onDrop = useCallback(async (acceptedFiles) => {
    if (acceptedFiles.length === 0) return;

    const file = acceptedFiles[0];

    if (!isSupported(file)) {
      setError(`Unsupported file format: ${file.name.split('.').pop()}`);
      return;
    }

    setView('loading');
    setError(null);

    try {
      const result = await parseDocument(file);
      setWords(result.words);
      setDocument({
        name: file.name,
        title: result.title,
        totalWords: result.words.length,
      });

      // Create/update session for persistence
      const hash = await createOrUpdateSession(
        result.words.slice(0, 1000).join(' '), // Use first 1000 words for hashing
        result.title,
        result.words.length,
        0,
        getSettings().defaultWpm || 300
      );
      setDocumentHash(hash);

      setView('reading');
    } catch (err) {
      console.error('Document parsing error:', err);
      setError(`Failed to parse document: ${err.message}`);
      setView('home');
    }
  }, []);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'text/plain': ['.txt'],
      'text/markdown': ['.md', '.markdown'],
      'application/pdf': ['.pdf'],
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx'],
      'application/epub+zip': ['.epub'],
      'application/vnd.oasis.opendocument.text': ['.odt'],
    },
    multiple: false,
  });

  const handleSavePosition = useCallback((position, wpm) => {
    if (documentHash && document) {
      saveSession(documentHash, {
        title: document.title,
        totalWords: document.totalWords,
        position,
        wpm,
        progress: document.totalWords > 0 ? (position / document.totalWords) * 100 : 0,
      });
    }
  }, [documentHash, document]);

  const handleExit = useCallback(() => {
    setView('home');
    setDocument(null);
    setWords([]);
    setDocumentHash(null);
  }, []);

  const handleDeleteSession = useCallback((hash) => {
    deleteSession(hash);
    setRecentSessions(getRecentSessions(5));
    toast({
      title: 'Session deleted',
      status: 'info',
      duration: 2000,
    });
  }, [toast]);

  // Render based on current view
  if (view === 'loading') {
    return (
      <Box minH="100vh" display="flex" alignItems="center" justifyContent="center">
        <VStack spacing={4}>
          <Spinner size="xl" color="brand.500" thickness="4px" />
          <Text color="gray.400">Parsing document...</Text>
        </VStack>
      </Box>
    );
  }

  if (view === 'reading') {
    return (
      <ReadingScreen
        words={words}
        documentName={document?.title || document?.name || 'Document'}
        onExit={handleExit}
        adsenseEnabled={adsenseConfig.enabled}
        adsenseKey={adsenseConfig.key}
        initialPosition={0}
        initialWpm={getSettings().defaultWpm || 300}
        onSavePosition={handleSavePosition}
      />
    );
  }

  // Home view
  return (
    <Box minH="100vh" display="flex" flexDirection="column">
      {/* Header */}
      <Box as="header" py={4} borderBottom="1px" borderColor="whiteAlpha.200">
        <Container maxW="container.xl">
          <HStack justify="space-between">
            <HStack spacing={3}>
              <Text fontSize="3xl">🐆</Text>
              <Heading size="lg">Cheetah</Heading>
            </HStack>
            <HStack spacing={4}>
              <Link
                href="/docs/"
                color="gray.400"
                _hover={{ color: 'brand.400' }}
                fontSize="sm"
              >
                Documentation
              </Link>
              <Text color="gray.500">|</Text>
              <Text color="gray.400" fontSize="sm">RSVP Speed Reading</Text>
            </HStack>
          </HStack>
        </Container>
      </Box>

      {/* Main Content */}
      <Box flex="1" display="flex" alignItems="center" justifyContent="center" py={8}>
        <Container maxW="container.lg">
          <VStack spacing={10}>
            {/* Hero Section */}
            <VStack spacing={4} textAlign="center">
              <Heading size="2xl" bgGradient="linear(to-r, brand.400, orange.300)" bgClip="text">
                Read at 1000+ WPM
              </Heading>
              <Text fontSize="xl" color="gray.400" maxW="600px">
                RSVP (Rapid Serial Visual Presentation) displays words one at a time,
                eliminating eye movement and unlocking your brain's true reading speed.
              </Text>
            </VStack>

            {/* Error Alert */}
            <AnimatePresence>
              {error && (
                <MotionBox
                  initial={{ opacity: 0, y: -20 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -20 }}
                  w="100%"
                >
                  <Alert status="error" borderRadius="lg">
                    <AlertIcon />
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                    <CloseButton
                      position="absolute"
                      right="8px"
                      top="8px"
                      onClick={() => setError(null)}
                    />
                  </Alert>
                </MotionBox>
              )}
            </AnimatePresence>

            {/* Drop Zone */}
            <Box
              {...getRootProps()}
              w="100%"
              maxW="600px"
              h="220px"
              border="3px dashed"
              borderColor={isDragActive ? 'brand.500' : 'whiteAlpha.300'}
              borderRadius="2xl"
              display="flex"
              alignItems="center"
              justifyContent="center"
              cursor="pointer"
              transition="all 0.3s"
              bg={isDragActive ? 'whiteAlpha.100' : 'transparent'}
              _hover={{ borderColor: 'brand.400', bg: 'whiteAlpha.50' }}
            >
              <input {...getInputProps()} />
              <VStack spacing={3}>
                <Text fontSize="5xl">📄</Text>
                <Text fontSize="xl" fontWeight="600">
                  {isDragActive
                    ? 'Drop your document here...'
                    : 'Drag & drop a document to start'}
                </Text>
                <Text fontSize="sm" color="gray.500">
                  or click to select a file
                </Text>
                <HStack spacing={2} flexWrap="wrap" justify="center">
                  {['PDF', 'DOCX', 'EPUB', 'ODT', 'TXT', 'MD'].map(format => (
                    <Badge key={format} colorScheme="gray" variant="subtle" fontSize="xs">
                      {format}
                    </Badge>
                  ))}
                </HStack>
              </VStack>
            </Box>

            {/* Recent Sessions */}
            {recentSessions.length > 0 && (
              <Box w="100%" maxW="800px">
                <Text fontSize="lg" fontWeight="600" mb={4} color="gray.300">
                  Continue Reading
                </Text>
                <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
                  {recentSessions.map(session => (
                    <Card
                      key={session.hash}
                      bg="whiteAlpha.50"
                      borderRadius="xl"
                      border="1px solid"
                      borderColor="whiteAlpha.100"
                      _hover={{ borderColor: 'brand.500', bg: 'whiteAlpha.100' }}
                      transition="all 0.2s"
                    >
                      <CardBody>
                        <HStack justify="space-between" align="start">
                          <VStack align="start" spacing={1} flex="1">
                            <Text fontWeight="600" noOfLines={1}>
                              {session.title}
                            </Text>
                            <HStack spacing={3} fontSize="sm" color="gray.500">
                              <Text>{session.totalWords?.toLocaleString()} words</Text>
                              <Text>•</Text>
                              <Text>{Math.round(session.progress || 0)}% complete</Text>
                            </HStack>
                            <HStack spacing={2}>
                              <Badge colorScheme="orange" fontSize="xs">
                                {session.wpm || 300} WPM
                              </Badge>
                              <Text fontSize="xs" color="gray.600">
                                {new Date(session.lastAccessed).toLocaleDateString()}
                              </Text>
                            </HStack>
                          </VStack>
                          <Tooltip label="Remove from history">
                            <IconButton
                              size="sm"
                              variant="ghost"
                              colorScheme="red"
                              icon={<Text>✕</Text>}
                              onClick={(e) => {
                                e.stopPropagation();
                                handleDeleteSession(session.hash);
                              }}
                            />
                          </Tooltip>
                        </HStack>
                      </CardBody>
                    </Card>
                  ))}
                </SimpleGrid>
                <Text fontSize="xs" color="gray.600" mt={3} textAlign="center">
                  Note: To resume, drop the same document again. Your position is saved automatically.
                </Text>
              </Box>
            )}

            {/* Features */}
            <SimpleGrid columns={{ base: 1, md: 3 }} spacing={6} w="100%" maxW="900px" mt={4}>
              <VStack p={6} bg="whiteAlpha.50" borderRadius="xl" spacing={3}>
                <Text fontSize="3xl">⚡</Text>
                <Text fontWeight="600">Lightning Fast</Text>
                <Text fontSize="sm" color="gray.400" textAlign="center">
                  Read 2-5x faster than traditional reading. Your brain processes words instantly without eye movement.
                </Text>
              </VStack>
              <VStack p={6} bg="whiteAlpha.50" borderRadius="xl" spacing={3}>
                <Text fontSize="3xl">🔒</Text>
                <Text fontWeight="600">Privacy First</Text>
                <Text fontSize="sm" color="gray.400" textAlign="center">
                  Documents are parsed entirely in your browser. Nothing is uploaded to any server.
                </Text>
              </VStack>
              <VStack p={6} bg="whiteAlpha.50" borderRadius="xl" spacing={3}>
                <Text fontSize="3xl">💾</Text>
                <Text fontWeight="600">Auto-Save</Text>
                <Text fontSize="sm" color="gray.400" textAlign="center">
                  Your reading position and speed preferences are saved automatically for each document.
                </Text>
              </VStack>
            </SimpleGrid>
          </VStack>
        </Container>
      </Box>

      {/* AdSense (if enabled) */}
      {adsenseConfig.enabled && (
        <Box py={4}>
          <AdSense publisherId={adsenseConfig.key} showPreview={true} />
        </Box>
      )}

      {/* Footer */}
      <Box as="footer" py={6} borderTop="1px" borderColor="whiteAlpha.200">
        <Container maxW="container.xl">
          <VStack spacing={4}>
            <HStack justify="center" spacing={4} flexWrap="wrap">
              <Text color="gray.500">
                Made with ❤️ by{' '}
                <Link href="https://kartoza.com" color="brand.400" isExternal>
                  Kartoza
                </Link>
              </Text>
              <Text color="gray.600">|</Text>
              <Link href="https://github.com/sponsors/timlinux" color="brand.400" isExternal>
                Donate!
              </Link>
              <Text color="gray.600">|</Text>
              <Link href="https://github.com/timlinux/cheetah" color="gray.400" isExternal>
                GitHub
              </Link>
            </HStack>
            <HStack spacing={4} fontSize="sm" color="gray.600">
              <Text>SPACE: pause/resume</Text>
              <Text>j/k: adjust speed</Text>
              <Text>1-9: speed presets</Text>
            </HStack>
          </VStack>
        </Container>
      </Box>
    </Box>
  );
}

export default App;
