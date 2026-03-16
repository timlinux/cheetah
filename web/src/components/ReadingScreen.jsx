// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import React, { useEffect, useState, useRef, useCallback } from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Flex,
  Link,
  Slider,
  SliderTrack,
  SliderFilledTrack,
  SliderThumb,
  Tooltip,
  Badge,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Input,
  Button,
  Switch,
  useDisclosure,
} from '@chakra-ui/react';
import AdSense from './AdSense.jsx';
import { getSettings, saveSettings } from '../storage.js';

// WPM Speed display bar
function WpmBar({ wpm, onWpmChange }) {
  const [showTooltip, setShowTooltip] = useState(false);

  const getColor = (wpm) => {
    if (wpm < 200) return 'blue.400';
    if (wpm < 350) return 'green.400';
    if (wpm < 500) return 'yellow.400';
    if (wpm < 700) return 'orange.400';
    return 'red.400';
  };

  const getLabel = (wpm) => {
    if (wpm < 200) return 'Relaxed';
    if (wpm < 350) return 'Normal';
    if (wpm < 500) return 'Fast';
    if (wpm < 700) return 'Very Fast';
    return 'Speed Demon';
  };

  return (
    <Box w="100%" maxW="500px" px={4}>
      <HStack justify="space-between" mb={2}>
        <Text color="gray.500" fontSize="sm">Reading Speed</Text>
        <Text color={getColor(wpm)} fontSize="sm" fontWeight="600">
          {getLabel(wpm)}
        </Text>
      </HStack>
      <Slider
        value={wpm}
        min={100}
        max={1000}
        step={25}
        onChange={onWpmChange}
        onMouseEnter={() => setShowTooltip(true)}
        onMouseLeave={() => setShowTooltip(false)}
      >
        <SliderTrack bg="gray.700" h="12px" borderRadius="full">
          <SliderFilledTrack
            bgGradient="linear(to-r, blue.500, green.400, yellow.400, orange.400, red.500)"
          />
        </SliderTrack>
        <Tooltip
          hasArrow
          bg="gray.700"
          color="white"
          placement="top"
          isOpen={showTooltip}
          label={`${wpm} WPM`}
        >
          <SliderThumb boxSize={6} bg="white" />
        </Tooltip>
      </Slider>
      <HStack justify="space-between" mt={1}>
        <Text color="gray.600" fontSize="xs">100</Text>
        <Text color="gray.600" fontSize="xs">550</Text>
        <Text color="gray.600" fontSize="xs">1000</Text>
      </HStack>
      <Flex justify="center" mt={4}>
        <VStack spacing={0}>
          <Text fontSize="5xl" fontWeight="700" color={getColor(wpm)}>
            {wpm}
          </Text>
          <Text fontSize="sm" color="gray.500">WPM</Text>
        </VStack>
      </Flex>
    </Box>
  );
}

// Speed preset buttons
function SpeedPresets({ onSelect, currentWpm }) {
  const presets = [
    { label: '1', wpm: 200 },
    { label: '2', wpm: 300 },
    { label: '3', wpm: 400 },
    { label: '4', wpm: 500 },
    { label: '5', wpm: 600 },
    { label: '6', wpm: 700 },
    { label: '7', wpm: 800 },
    { label: '8', wpm: 900 },
    { label: '9', wpm: 1000 },
  ];

  return (
    <HStack spacing={1} justify="center" flexWrap="wrap">
      {presets.map(({ label, wpm }) => (
        <Box
          key={label}
          as="button"
          px={3}
          py={1}
          fontSize="sm"
          fontWeight="600"
          color={currentWpm === wpm ? 'white' : 'gray.400'}
          bg={currentWpm === wpm ? 'brand.500' : 'transparent'}
          borderRadius="md"
          border="1px solid"
          borderColor={currentWpm === wpm ? 'brand.500' : 'gray.600'}
          _hover={{ bg: 'whiteAlpha.100', borderColor: 'brand.400' }}
          onClick={() => onSelect(wpm)}
          transition="all 0.15s"
        >
          {label}
        </Box>
      ))}
    </HStack>
  );
}

// Heading breadcrumb display
function HeadingBreadcrumb({ headings, currentListNumber }) {
  const hasContent = headings.some(h => h) || currentListNumber;
  if (!hasContent) return null;

  return (
    <Box
      px={6}
      py={3}
      bg="whiteAlpha.50"
      borderBottom="1px solid"
      borderColor="whiteAlpha.100"
    >
      <VStack align="start" spacing={1}>
        {headings.map((heading, level) => {
          if (!heading) return null;
          const fontSize = ['xl', 'lg', 'md', 'sm', 'sm', 'xs'][level] || 'xs';
          const color = ['brand.400', 'brand.300', 'gray.300', 'gray.400', 'gray.500', 'gray.500'][level] || 'gray.500';
          const indent = level * 16;

          return (
            <Text
              key={level}
              fontSize={fontSize}
              fontWeight={level < 2 ? '600' : '500'}
              color={color}
              pl={`${indent}px`}
              fontFamily="Georgia, 'Times New Roman', serif"
            >
              {heading}
            </Text>
          );
        })}
        {currentListNumber && (
          <HStack pl={headings.filter(h => h).length * 16 + 'px'}>
            <Badge colorScheme="orange" fontSize="sm" px={2}>
              {currentListNumber}.
            </Badge>
          </HStack>
        )}
      </VStack>
    </Box>
  );
}

function ReadingScreen({
  words,
  documentName,
  onExit,
  adsenseEnabled,
  adsenseKey,
  initialPosition = 0,
  initialWpm = 300,
  onSavePosition,
}) {
  const [currentIndex, setCurrentIndex] = useState(initialPosition);
  const [wpm, setWpm] = useState(initialWpm);
  const [isPaused, setIsPaused] = useState(true);
  // Headings state: array of 6 levels [h1, h2, h3, h4, h5, h6]
  const [headings, setHeadings] = useState(['', '', '', '', '', '']);
  const [currentListNumber, setCurrentListNumber] = useState(null);
  const timerRef = useRef(null);

  // All caps display setting (default: on)
  const [displayAllCaps, setDisplayAllCaps] = useState(() => {
    const settings = getSettings();
    return settings.displayAllCaps !== false; // Default to true if not set
  });

  // Go-to modal state
  const { isOpen: isGotoOpen, onOpen: onGotoOpen, onClose: onGotoClose } = useDisclosure();
  const [gotoInput, setGotoInput] = useState('');
  const [wasPausedBeforeGoto, setWasPausedBeforeGoto] = useState(true);
  const gotoInputRef = useRef(null);

  // Get current word (handle both string and object formats)
  const currentWordObj = words[currentIndex] || {};
  const currentWord = typeof currentWordObj === 'string' ? currentWordObj : currentWordObj.text || '';

  const progress = words.length > 0 ? ((currentIndex + 1) / words.length) * 100 : 0;

  // Collect full heading text from consecutive heading words
  const collectHeadingText = useCallback((startIndex, level) => {
    let text = '';
    for (let i = startIndex; i < words.length; i++) {
      const w = words[i];
      if (typeof w === 'string') break;
      if (w.headingLevel !== level) break;
      text += (text ? ' ' : '') + w.text;
      if (w.isHeadingEnd) break;
    }
    return text;
  }, [words]);

  // Update headings and list state based on current word
  useEffect(() => {
    const wordObj = words[currentIndex];
    if (!wordObj || typeof wordObj === 'string') {
      // Plain text without metadata - check if we should clear list
      setCurrentListNumber(null);
      return;
    }

    // Handle headings
    if (wordObj.headingLevel && wordObj.isHeadingStart) {
      const level = wordObj.headingLevel - 1; // 0-indexed
      const headingText = collectHeadingText(currentIndex, wordObj.headingLevel);

      setHeadings(prev => {
        const newHeadings = [...prev];
        // Set this level
        newHeadings[level] = headingText;
        // Clear all lower levels (higher numbers)
        for (let i = level + 1; i < 6; i++) {
          newHeadings[i] = '';
        }
        return newHeadings;
      });
      setCurrentListNumber(null); // New heading clears list context
    }

    // Handle numbered lists
    if (wordObj.listNumber !== undefined) {
      if (wordObj.isListStart) {
        setCurrentListNumber(wordObj.listNumber);
      }
    } else if (!wordObj.headingLevel && !wordObj.isBullet) {
      // Regular text without list markers - might be end of list
      // Only clear if previous words were list items and now we're not
      const prevWord = words[currentIndex - 1];
      if (prevWord && (prevWord.listNumber !== undefined || prevWord.isListEnd)) {
        // Keep list number visible until we leave the list item content
      }
    }

    // Clear list number when we hit a new list item or non-list content
    if (wordObj.isListStart && wordObj.listNumber !== currentListNumber) {
      setCurrentListNumber(wordObj.listNumber);
    }
  }, [currentIndex, words, collectHeadingText, currentListNumber]);

  // Calculate delay based on WPM and content type
  const getWordDelay = useCallback((wordObj) => {
    const baseDelay = 60000 / wpm;
    const word = typeof wordObj === 'string' ? wordObj : wordObj.text || '';

    // Headings get extra pause
    if (typeof wordObj !== 'string' && wordObj.headingLevel) {
      if (wordObj.isHeadingEnd) {
        return baseDelay * 3; // Pause at end of heading
      }
      return baseDelay * 1.5;
    }

    // List item start gets a pause
    if (typeof wordObj !== 'string' && wordObj.isListStart) {
      return baseDelay * 1.5;
    }

    // Punctuation pauses
    if (word.match(/[.!?]$/)) {
      return baseDelay * 2;
    }
    if (word.match(/[,;:]$/)) {
      return baseDelay * 1.3;
    }
    if (word.match(/[-—]$/)) {
      return baseDelay * 1.2;
    }

    return baseDelay;
  }, [wpm]);

  // Advance to next word
  const advanceWord = useCallback(() => {
    setCurrentIndex(prev => {
      if (prev >= words.length - 1) {
        setIsPaused(true);
        return prev;
      }
      return prev + 1;
    });
  }, [words.length]);

  // Timer effect
  useEffect(() => {
    if (isPaused || currentIndex >= words.length - 1) {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
      return;
    }

    const delay = getWordDelay(words[currentIndex]);
    timerRef.current = setTimeout(advanceWord, delay);

    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, [isPaused, currentIndex, words, getWordDelay, advanceWord]);

  // Save position periodically
  useEffect(() => {
    const saveInterval = setInterval(() => {
      if (onSavePosition) {
        onSavePosition(currentIndex, wpm);
      }
    }, 5000);

    return () => clearInterval(saveInterval);
  }, [currentIndex, wpm, onSavePosition]);

  // Return to start
  const returnToStart = useCallback(() => {
    setCurrentIndex(0);
    setHeadings(['', '', '', '', '', '']);
    setCurrentListNumber(null);
    setIsPaused(true);
  }, []);

  // Toggle all caps display
  const toggleAllCaps = useCallback(() => {
    setDisplayAllCaps(prev => {
      const newValue = !prev;
      saveSettings({ displayAllCaps: newValue });
      return newValue;
    });
  }, []);

  // Open go-to modal
  const openGotoModal = useCallback(() => {
    setWasPausedBeforeGoto(isPaused);
    setIsPaused(true);
    setGotoInput('');
    onGotoOpen();
    // Focus input after modal opens
    setTimeout(() => {
      if (gotoInputRef.current) {
        gotoInputRef.current.focus();
      }
    }, 100);
  }, [isPaused, onGotoOpen]);

  // Handle go-to submit
  const handleGotoSubmit = useCallback(() => {
    const percentage = parseFloat(gotoInput);
    if (!isNaN(percentage) && percentage >= 0 && percentage <= 100) {
      const targetIndex = Math.floor((percentage / 100) * words.length);
      setCurrentIndex(Math.min(Math.max(0, targetIndex), words.length - 1));
    }
    onGotoClose();
    if (!wasPausedBeforeGoto) {
      setIsPaused(false);
    }
  }, [gotoInput, words.length, onGotoClose, wasPausedBeforeGoto]);

  // Handle go-to cancel
  const handleGotoCancel = useCallback(() => {
    onGotoClose();
    if (!wasPausedBeforeGoto) {
      setIsPaused(false);
    }
  }, [onGotoClose, wasPausedBeforeGoto]);

  // Keyboard controls
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.ctrlKey || e.metaKey || e.altKey) return;

      switch (e.key) {
        case ' ':
          e.preventDefault();
          setIsPaused(prev => !prev);
          break;
        case 'Escape':
        case 'b':
        case 'B':
          e.preventDefault();
          if (onSavePosition) {
            onSavePosition(currentIndex, wpm);
          }
          onExit();
          break;
        case 'r':
        case 'R':
          e.preventDefault();
          returnToStart();
          break;
        case 'j':
        case 'ArrowLeft':
          e.preventDefault();
          setWpm(prev => Math.max(100, prev - 50));
          break;
        case 'k':
        case 'ArrowRight':
          e.preventDefault();
          setWpm(prev => Math.min(1000, prev + 50));
          break;
        case 'h':
        case 'ArrowUp':
          e.preventDefault();
          setCurrentIndex(prev => {
            for (let i = prev - 2; i >= 0; i--) {
              const w = words[i];
              const text = typeof w === 'string' ? w : w.text || '';
              if (text.match(/[.!?]$/)) {
                return i + 1;
              }
            }
            return 0;
          });
          break;
        case 'l':
        case 'ArrowDown':
          e.preventDefault();
          setCurrentIndex(prev => {
            for (let i = prev + 1; i < words.length; i++) {
              const w = words[i];
              const text = typeof w === 'string' ? w : w.text || '';
              if (text.match(/[.!?]$/)) {
                return Math.min(i + 1, words.length - 1);
              }
            }
            return words.length - 1;
          });
          break;
        case '1':
        case '2':
        case '3':
        case '4':
        case '5':
        case '6':
        case '7':
        case '8':
        case '9':
          e.preventDefault();
          const presetWpm = [200, 300, 400, 500, 600, 700, 800, 900, 1000][parseInt(e.key) - 1];
          setWpm(presetWpm);
          break;
        case 'g':
        case 'G':
          e.preventDefault();
          if (!isGotoOpen) {
            openGotoModal();
          }
          break;
        case 'c':
        case 'C':
          e.preventDefault();
          toggleAllCaps();
          break;
        default:
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [currentIndex, wpm, words, onExit, onSavePosition, returnToStart, isGotoOpen, openGotoModal, toggleAllCaps]);

  return (
    <Flex minH="100vh" direction="column" overflow="hidden">
      {/* Header */}
      <Flex justify="space-between" align="center" px={6} py={4} borderBottom="1px solid" borderColor="whiteAlpha.100">
        <HStack spacing={4}>
          <Text fontSize="xl" fontWeight="bold" color="brand.400">
            Cheetah
          </Text>
          <Link
            href="/docs/"
            color="gray.500"
            fontSize="xs"
            _hover={{ color: 'brand.400' }}
          >
            Docs
          </Link>
          <Text color="gray.600">|</Text>
          <Text color="gray.500" fontSize="sm" noOfLines={1} maxW="300px">
            {documentName}
          </Text>
        </HStack>
        <HStack spacing={3}>
          <Box
            as="button"
            px={4}
            py={2}
            h="40px"
            bg="gray.700"
            color="gray.300"
            borderRadius="lg"
            fontSize="sm"
            fontWeight="600"
            onClick={onExit}
            _hover={{ bg: 'gray.600' }}
            display="flex"
            alignItems="center"
          >
            ← Back
          </Box>
          <Box
            as="button"
            px={4}
            py={2}
            h="40px"
            bg="gray.700"
            color="gray.300"
            borderRadius="lg"
            fontSize="sm"
            fontWeight="600"
            onClick={returnToStart}
            _hover={{ bg: 'gray.600' }}
            display="flex"
            alignItems="center"
          >
            ↺ Restart
          </Box>
          <HStack spacing={2} px={3} py={2} bg="gray.700" borderRadius="lg" h="40px">
            <Text fontSize="xs" color="gray.400">CAPS</Text>
            <Switch
              size="sm"
              colorScheme="brand"
              isChecked={displayAllCaps}
              onChange={toggleAllCaps}
            />
          </HStack>
          <VStack spacing={0} minW="100px">
            <Text color="gray.400" fontSize="xs">Word</Text>
            <Text color="white" fontSize="lg" fontWeight="bold">
              {currentIndex + 1} / {words.length}
            </Text>
          </VStack>
          <Box
            as="button"
            px={4}
            py={2}
            h="40px"
            bg={isPaused ? 'green.500' : 'orange.500'}
            color="white"
            borderRadius="lg"
            fontWeight="600"
            onClick={() => setIsPaused(prev => !prev)}
            _hover={{ opacity: 0.9 }}
            transition="all 0.15s"
            display="flex"
            alignItems="center"
          >
            {isPaused ? '▶ Play' : '⏸ Pause'}
          </Box>
        </HStack>
      </Flex>

      {/* Interactive progress scrubber */}
      <Box px={6} py={2}>
        <Slider
          value={currentIndex}
          min={0}
          max={Math.max(0, words.length - 1)}
          onChange={(val) => {
            setIsPaused(true); // Pause while scrubbing
            setCurrentIndex(val);
          }}
          onChangeStart={() => {
            setIsPaused(true); // Pause when starting to drag
          }}
          focusThumbOnChange={false}
        >
          <SliderTrack bg="gray.700" h="12px" borderRadius="full">
            <SliderFilledTrack
              bgGradient="linear(to-r, brand.500, green.400)"
              transition="width 0.05s linear"
            />
          </SliderTrack>
          <Tooltip
            hasArrow
            bg="gray.700"
            color="white"
            placement="top"
            label={`Word ${currentIndex + 1} / ${words.length}`}
          >
            <SliderThumb boxSize={5} bg="white" _focus={{ boxShadow: 'outline' }} />
          </Tooltip>
        </Slider>
        <HStack justify="space-between" mt={1}>
          <Text fontSize="xs" color="gray.500">
            Word {currentIndex + 1} / {words.length}
          </Text>
          <Text fontSize="xs" color="gray.500">
            {progress.toFixed(1)}% complete
          </Text>
        </HStack>
      </Box>

      {/* Heading breadcrumb trail */}
      <HeadingBreadcrumb headings={headings} currentListNumber={currentListNumber} />

      {/* Main reading area */}
      <Flex flex={1} direction="column" align="center" justify="center" position="relative">
        <Box
          position="relative"
          minH="200px"
          display="flex"
          alignItems="center"
          justifyContent="center"
          px={4}
        >
          <Text
            fontSize={{ base: '6xl', md: '8xl', lg: '9xl' }}
            fontWeight="400"
            fontFamily="Georgia, 'Times New Roman', 'Noto Serif', serif"
            color="white"
            textAlign="center"
            letterSpacing="0.01em"
            lineHeight="1.1"
            userSelect="none"
            textTransform={displayAllCaps ? 'uppercase' : 'none'}
          >
            {currentWord}
          </Text>
        </Box>

        {isPaused && (
          <Box position="absolute" bottom="15%">
            <Text
              fontSize="lg"
              color="orange.400"
              fontWeight="600"
              bg="rgba(0,0,0,0.6)"
              px={5}
              py={2}
              borderRadius="lg"
            >
              PAUSED - Press SPACE to resume
            </Text>
          </Box>
        )}
      </Flex>

      {/* Speed controls */}
      <Box pb={4}>
        <Flex justify="center" mb={4}>
          <WpmBar wpm={wpm} onWpmChange={setWpm} />
        </Flex>
        <SpeedPresets onSelect={setWpm} currentWpm={wpm} />
      </Box>

      {/* Keyboard hints */}
      <Flex justify="center" pb={3}>
        <HStack spacing={3} color="gray.600" fontSize="sm" flexWrap="wrap" justify="center">
          <Text><b>SPACE</b> pause</Text>
          <Text>|</Text>
          <Text><b>r</b> restart</Text>
          <Text>|</Text>
          <Text><b>j/k</b> speed</Text>
          <Text>|</Text>
          <Text><b>h/l</b> paragraph</Text>
          <Text>|</Text>
          <Text><b>1-9</b> presets</Text>
          <Text>|</Text>
          <Text><b>g</b> goto</Text>
          <Text>|</Text>
          <Text><b>c</b> caps</Text>
          <Text>|</Text>
          <Text><b>ESC/b</b> back</Text>
        </HStack>
      </Flex>

      {/* Footer */}
      <Flex justify="center" pb={3}>
        <HStack spacing={2} color="gray.600" fontSize="xs">
          <Text>Made with</Text>
          <Text color="red.400">♥</Text>
          <Text>by</Text>
          <Link href="https://kartoza.com" isExternal color="brand.500" _hover={{ color: 'brand.400' }}>
            Kartoza
          </Link>
          <Text>|</Text>
          <Link href="https://github.com/sponsors/timlinux" isExternal color="blue.400" _hover={{ color: 'blue.300' }}>
            Donate!
          </Link>
          <Text>|</Text>
          <Link href="https://github.com/timlinux/cheetah" isExternal color="gray.500" _hover={{ color: 'gray.400' }}>
            GitHub
          </Link>
        </HStack>
      </Flex>

      {adsenseEnabled && (
        <Box pb={4}>
          <AdSense publisherId={adsenseKey} showPreview={true} />
        </Box>
      )}

      {/* Go-to Position Modal */}
      <Modal isOpen={isGotoOpen} onClose={handleGotoCancel} isCentered>
        <ModalOverlay bg="blackAlpha.700" />
        <ModalContent bg="gray.800" borderColor="brand.500" borderWidth="1px">
          <ModalHeader color="brand.400">Go to Position</ModalHeader>
          <ModalBody>
            <VStack spacing={4}>
              <Text color="gray.300">Enter percentage (0-100):</Text>
              <Input
                ref={gotoInputRef}
                type="number"
                min={0}
                max={100}
                step={0.1}
                value={gotoInput}
                onChange={(e) => setGotoInput(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    handleGotoSubmit();
                  } else if (e.key === 'Escape') {
                    e.preventDefault();
                    handleGotoCancel();
                  }
                }}
                placeholder="50"
                size="lg"
                textAlign="center"
                fontSize="2xl"
                bg="gray.700"
                borderColor="gray.600"
                _hover={{ borderColor: 'brand.400' }}
                _focus={{ borderColor: 'brand.500', boxShadow: '0 0 0 1px var(--chakra-colors-brand-500)' }}
              />
              {gotoInput && !isNaN(parseFloat(gotoInput)) && (
                <Text color="gray.400" fontSize="sm">
                  → Word {Math.min(Math.max(0, Math.floor((parseFloat(gotoInput) / 100) * words.length)), words.length - 1) + 1} / {words.length}
                </Text>
              )}
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={handleGotoCancel}>
              Cancel
            </Button>
            <Button colorScheme="brand" onClick={handleGotoSubmit}>
              Go
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Flex>
  );
}

export default ReadingScreen;
