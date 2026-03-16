// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import Foundation
import PDFKit
import zlib

// MARK: - DocumentParser

/// Multi-format document parser.
/// Supported formats: TXT, Markdown, PDF (PDFKit), DOCX, EPUB, ODT (ZIP+XML).
enum DocumentParser {

    // MARK: - Errors

    enum ParseError: LocalizedError {
        case unsupportedFormat(String)
        case readFailed(String)
        case parseFailed(String)

        var errorDescription: String? {
            switch self {
            case .unsupportedFormat(let ext): return "Unsupported format: \(ext)"
            case .readFailed(let msg):        return "Failed to read file: \(msg)"
            case .parseFailed(let msg):       return "Failed to parse document: \(msg)"
            }
        }
    }

    // MARK: - Supported extensions

    static let supportedExtensions: Set<String> = ["txt", "md", "markdown", "pdf", "docx", "epub", "odt"]

    static func isSupported(url: URL) -> Bool {
        supportedExtensions.contains(url.pathExtension.lowercased())
    }

    // MARK: - Public entry point

    /// Parse a document at the given URL.
    static func parse(url: URL) throws -> Document {
        let ext = url.pathExtension.lowercased()
        switch ext {
        case "txt", "md", "markdown":
            return try parsePlainText(url: url)
        case "pdf":
            return try parsePDF(url: url)
        case "docx":
            let data = try Data(contentsOf: url)
            return try parseDocx(data: data, title: url.deletingPathExtension().lastPathComponent)
        case "epub":
            let data = try Data(contentsOf: url)
            return try parseEpub(data: data, title: url.deletingPathExtension().lastPathComponent)
        case "odt":
            let data = try Data(contentsOf: url)
            return try parseOdt(data: data, title: url.deletingPathExtension().lastPathComponent)
        default:
            throw ParseError.unsupportedFormat(ext)
        }
    }

    // MARK: - TXT / Markdown

    private static func parsePlainText(url: URL) throws -> Document {
        guard let text = (try? String(contentsOf: url, encoding: .utf8)) ??
                         (try? String(contentsOf: url, encoding: .isoLatin1)) else {
            throw ParseError.readFailed(url.lastPathComponent)
        }
        let title = url.deletingPathExtension().lastPathComponent
        return buildDocument(title: title, path: url.path, rawText: text)
    }

    // MARK: - PDF (PDFKit)

    private static func parsePDF(url: URL) throws -> Document {
        guard let pdf = PDFDocument(url: url) else {
            throw ParseError.parseFailed("Could not open PDF: \(url.lastPathComponent)")
        }

        var allText = ""
        for pageIndex in 0..<pdf.pageCount {
            guard let page = pdf.page(at: pageIndex),
                  let text = page.string else { continue }
            allText += text + "\n\n"
        }

        let title = (pdf.documentAttributes?[PDFDocumentAttribute.titleAttribute] as? String)
            ?? url.deletingPathExtension().lastPathComponent

        return buildDocument(title: title, path: url.path, rawText: allText)
    }

    // MARK: - DOCX

    private static func parseDocx(data: Data, title: String) throws -> Document {
        guard let entries = try? ZipReader.entries(from: data),
              let xmlEntry = entries.first(where: { $0.filename == "word/document.xml" }),
              let xmlData = try? ZipReader.extract(entry: xmlEntry, from: data) else {
            throw ParseError.parseFailed("Could not read DOCX archive or locate word/document.xml")
        }

        // Extract text from Word XML: paragraphs are <w:p>, text runs are <w:t>
        let extractor = WordXMLExtractor()
        extractor.parse(xmlData)
        let rawText = extractor.paragraphs.joined(separator: "\n\n")
        return buildDocument(title: title, path: nil, rawText: rawText)
    }

    // MARK: - EPUB

    private static func parseEpub(data: Data, title: String) throws -> Document {
        guard let entries = try? ZipReader.entries(from: data) else {
            throw ParseError.parseFailed("Could not read EPUB archive")
        }

        var docTitle = title
        var orderedPaths: [String] = []

        // Find and parse OPF for spine order + metadata
        if let opfEntry = entries.first(where: { $0.filename.hasSuffix(".opf") }),
           let opfData = try? ZipReader.extract(entry: opfEntry, from: data) {
            let opfParser = OPFParser(opfPath: opfEntry.filename)
            opfParser.parse(opfData)
            orderedPaths = opfParser.orderedContentPaths
            if let t = opfParser.title, !t.isEmpty { docTitle = t }
        }

        // Fallback to all HTML files
        if orderedPaths.isEmpty {
            orderedPaths = entries
                .filter { $0.filename.hasSuffix(".html") || $0.filename.hasSuffix(".xhtml") }
                .map { $0.filename }
        }

        var allText = ""
        for path in orderedPaths {
            if let entry = entries.first(where: { $0.filename == path }),
               let htmlData = try? ZipReader.extract(entry: entry, from: data) {
                allText += HTMLTextExtractor.extract(from: htmlData) + "\n\n"
            }
        }

        return buildDocument(title: docTitle, path: nil, rawText: allText)
    }

    // MARK: - ODT

    private static func parseOdt(data: Data, title: String) throws -> Document {
        guard let entries = try? ZipReader.entries(from: data),
              let xmlEntry = entries.first(where: { $0.filename == "content.xml" }),
              let xmlData = try? ZipReader.extract(entry: xmlEntry, from: data) else {
            throw ParseError.parseFailed("Could not read ODT archive or locate content.xml")
        }

        let extractor = ODTXMLExtractor()
        extractor.parse(xmlData)
        let rawText = extractor.paragraphs.joined(separator: "\n\n")
        return buildDocument(title: title, path: nil, rawText: rawText)
    }

    // MARK: - Text processing helpers

    /// Build a Document from a raw text string by splitting into paragraphs.
    private static func buildDocument(title: String, path: String?, rawText: String) -> Document {
        let paragraphTexts = splitIntoParagraphs(rawText)
        let paragraphWords = paragraphTexts.map { tokenize($0) }
        return Document.build(title: title, path: path, paragraphWords: paragraphWords)
    }

    /// Split raw text into paragraphs on blank lines.
    static func splitIntoParagraphs(_ text: String) -> [String] {
        let normalised = text
            .replacingOccurrences(of: "\r\n", with: "\n")
            .replacingOccurrences(of: "\r", with: "\n")

        return normalised.components(separatedBy: "\n\n")
            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines)
                      .replacingOccurrences(of: "\n", with: " ") }
            .filter { !$0.isEmpty }
    }

    /// Tokenise a paragraph into words.
    static func tokenize(_ text: String) -> [String] {
        text.components(separatedBy: .whitespaces)
            .map { $0.trimmingCharacters(in: .init(charactersIn: "\u{FEFF}\u{00A0}")) }
            .filter { !$0.isEmpty }
    }
}

// MARK: - WordXMLExtractor

/// SAX-based parser for DOCX `word/document.xml`.
/// Collects text from `<w:t>` elements, grouped by `<w:p>` paragraphs.
private final class WordXMLExtractor: NSObject, XMLParserDelegate {

    private(set) var paragraphs: [String] = []
    private var currentParagraphText = ""
    private var inTextElement = false

    func parse(_ data: Data) {
        let parser = XMLParser(data: data)
        parser.delegate = self
        parser.parse()
    }

    func parser(_ parser: XMLParser,
                didStartElement elementName: String,
                namespaceURI: String?,
                qualifiedName qName: String?,
                attributes attributeDict: [String: String] = [:]) {
        // <w:p> starts a new paragraph
        if qName == "w:p" || (elementName == "p" && namespaceURI?.contains("wordprocessingml") == true) {
            currentParagraphText = ""
        }
        // <w:t> contains text
        if qName == "w:t" || (elementName == "t" && namespaceURI?.contains("wordprocessingml") == true) {
            inTextElement = true
        }
    }

    func parser(_ parser: XMLParser,
                didEndElement elementName: String,
                namespaceURI: String?,
                qualifiedName qName: String?) {
        if qName == "w:t" || (elementName == "t" && namespaceURI?.contains("wordprocessingml") == true) {
            inTextElement = false
        }
        if qName == "w:p" || (elementName == "p" && namespaceURI?.contains("wordprocessingml") == true) {
            let text = currentParagraphText.trimmingCharacters(in: .whitespaces)
            if !text.isEmpty { paragraphs.append(text) }
            currentParagraphText = ""
        }
    }

    func parser(_ parser: XMLParser, foundCharacters string: String) {
        if inTextElement { currentParagraphText += string }
    }
}

// MARK: - ODTXMLExtractor

/// SAX-based parser for ODT `content.xml`.
/// Collects text from `<text:p>` elements.
private final class ODTXMLExtractor: NSObject, XMLParserDelegate {

    private(set) var paragraphs: [String] = []
    private var currentText = ""
    private var depth = 0
    private var inParagraph = false

    func parse(_ data: Data) {
        let parser = XMLParser(data: data)
        parser.delegate = self
        parser.parse()
    }

    func parser(_ parser: XMLParser,
                didStartElement elementName: String,
                namespaceURI: String?,
                qualifiedName qName: String?,
                attributes attributeDict: [String: String] = [:]) {
        // text:p, text:h are paragraph-level elements in ODF
        if elementName == "p" || elementName == "h" {
            inParagraph = true
            currentText = ""
        }
    }

    func parser(_ parser: XMLParser,
                didEndElement elementName: String,
                namespaceURI: String?,
                qualifiedName qName: String?) {
        if elementName == "p" || elementName == "h" {
            let text = currentText.trimmingCharacters(in: .whitespaces)
            if !text.isEmpty { paragraphs.append(text) }
            currentText = ""
            inParagraph = false
        }
    }

    func parser(_ parser: XMLParser, foundCharacters string: String) {
        if inParagraph { currentText += string }
    }
}

// MARK: - HTMLTextExtractor

/// Extracts plain text from HTML data by stripping tags with XMLParser.
enum HTMLTextExtractor {

    static func extract(from data: Data) -> String {
        // First try NSAttributedString-based extraction (cleanest)
        if let text = attributedStringExtract(from: data) { return text }

        // Fallback: simple tag stripper
        let raw = String(data: data, encoding: .utf8)
                   ?? String(data: data, encoding: .isoLatin1)
                   ?? ""
        return stripTags(from: raw)
    }

    private static func attributedStringExtract(from data: Data) -> String? {
        // Use NSAttributedString HTML init (main-thread only; safe here as we call on background)
        guard let html = String(data: data, encoding: .utf8) else { return nil }
        // Remove script/style before extraction
        let cleaned = removeScriptStyle(from: html)
        let stripped = stripTags(from: cleaned)
        // Collapse extra whitespace
        return stripped.components(separatedBy: .newlines)
            .map { $0.trimmingCharacters(in: .whitespaces) }
            .filter { !$0.isEmpty }
            .joined(separator: "\n")
    }

    private static func removeScriptStyle(from html: String) -> String {
        var result = html
        for tag in ["script", "style"] {
            let open = "<\(tag)"
            let close = "</\(tag)>"
            while true {
                guard let startRange = result.range(of: open, options: .caseInsensitive) else { break }
                if let endRange = result.range(of: close, options: .caseInsensitive,
                                               range: startRange.lowerBound..<result.endIndex) {
                    result.removeSubrange(startRange.lowerBound..<endRange.upperBound)
                } else {
                    // No closing tag found – just remove from open tag to end
                    result.removeSubrange(startRange.lowerBound..<result.endIndex)
                    break
                }
            }
        }
        return result
    }

    static func stripTags(from text: String) -> String {
        var result = ""
        var inTag = false
        for ch in text {
            if ch == "<" { inTag = true }
            else if ch == ">" {
                inTag = false
                result += " "
            } else if !inTag {
                result.append(ch)
            }
        }
        return result
    }
}

// MARK: - OPFParser

/// SAX-based parser for EPUB OPF files.
/// Extracts the spine reading order and document title.
private final class OPFParser: NSObject, XMLParserDelegate {

    private(set) var title: String? = nil
    private(set) var orderedContentPaths: [String] = []

    private let opfDir: String
    private var manifest: [String: String] = [:]  // id -> href
    private var spineIdrefs: [String] = []
    private var inTitle = false
    private var titleText = ""

    init(opfPath: String) {
        let dir = (opfPath as NSString).deletingLastPathComponent
        self.opfDir = dir
    }

    func parse(_ data: Data) {
        let parser = XMLParser(data: data)
        parser.delegate = self
        parser.parse()

        // Resolve spine idrefs to paths
        orderedContentPaths = spineIdrefs.compactMap { idref -> String? in
            guard let href = manifest[idref] else { return nil }
            if opfDir.isEmpty { return href }
            return "\(opfDir)/\(href)"
        }
    }

    func parser(_ parser: XMLParser,
                didStartElement elementName: String,
                namespaceURI: String?,
                qualifiedName qName: String?,
                attributes attrs: [String: String] = [:]) {
        switch elementName {
        case "title":
            inTitle = true
            titleText = ""
        case "item":
            if let id = attrs["id"], let href = attrs["href"] {
                manifest[id] = href
            }
        case "itemref":
            if let idref = attrs["idref"] {
                spineIdrefs.append(idref)
            }
        default: break
        }
    }

    func parser(_ parser: XMLParser,
                didEndElement elementName: String,
                namespaceURI: String?,
                qualifiedName qName: String?) {
        if elementName == "title" {
            inTitle = false
            if title == nil && !titleText.trimmingCharacters(in: .whitespaces).isEmpty {
                title = titleText.trimmingCharacters(in: .whitespaces)
            }
        }
    }

    func parser(_ parser: XMLParser, foundCharacters string: String) {
        if inTitle { titleText += string }
    }
}

// MARK: - ZipReader

/// Minimal ZIP archive reader implemented in pure Swift + zlib.
/// Handles stored (method 0) and deflated (method 8) entries.
enum ZipReader {

    // MARK: - Entry descriptor

    struct Entry {
        let filename: String
        let compressionMethod: UInt16
        let compressedSize: UInt32
        let uncompressedSize: UInt32
        let localHeaderOffset: Int
    }

    // MARK: - Errors

    enum ZipError: Error {
        case invalidSignature
        case unsupportedCompression(UInt16)
        case decompressFailed
        case truncated
    }

    // MARK: - Parse entries from central directory

    /// Returns all entries listed in the ZIP central directory.
    static func entries(from data: Data) throws -> [Entry] {
        let bytes = [UInt8](data)
        let count = bytes.count

        // Locate End-of-Central-Directory (EOCD): signature PK\x05\x06
        guard count >= 22 else { throw ZipError.truncated }

        var eocdOffset = -1
        var i = count - 22
        while i >= 0 {
            if bytes[i] == 0x50 && bytes[i+1] == 0x4B &&
               bytes[i+2] == 0x05 && bytes[i+3] == 0x06 {
                eocdOffset = i
                break
            }
            if i == 0 { break }
            i -= 1
        }
        guard eocdOffset >= 0 else { throw ZipError.invalidSignature }

        let totalEntries = Int(readUInt16LE(bytes, at: eocdOffset + 10))
        let cdOffset     = Int(readUInt32LE(bytes, at: eocdOffset + 16))

        var entries: [Entry] = []
        var pos = cdOffset

        for _ in 0..<totalEntries {
            guard pos + 46 <= count,
                  bytes[pos] == 0x50 && bytes[pos+1] == 0x4B &&
                  bytes[pos+2] == 0x01 && bytes[pos+3] == 0x02 else { break }

            let compressionMethod = readUInt16LE(bytes, at: pos + 10)
            let compressedSize    = readUInt32LE(bytes, at: pos + 20)
            let uncompressedSize  = readUInt32LE(bytes, at: pos + 24)
            let fileNameLen       = Int(readUInt16LE(bytes, at: pos + 28))
            let extraLen          = Int(readUInt16LE(bytes, at: pos + 30))
            let commentLen        = Int(readUInt16LE(bytes, at: pos + 32))
            let localHeaderOffset = Int(readUInt32LE(bytes, at: pos + 42))

            guard pos + 46 + fileNameLen <= count else { break }
            let filenameBytes = Array(bytes[(pos + 46)..<(pos + 46 + fileNameLen)])
            let filename = String(bytes: filenameBytes, encoding: .utf8)
                        ?? String(bytes: filenameBytes, encoding: .isoLatin1)
                        ?? ""

            entries.append(Entry(
                filename: filename,
                compressionMethod: compressionMethod,
                compressedSize: compressedSize,
                uncompressedSize: uncompressedSize,
                localHeaderOffset: localHeaderOffset
            ))

            pos += 46 + fileNameLen + extraLen + commentLen
        }

        return entries
    }

    // MARK: - Extract a single entry

    /// Extracts and decompresses a single entry's data.
    static func extract(entry: Entry, from data: Data) throws -> Data {
        let bytes = [UInt8](data)
        let pos = entry.localHeaderOffset

        guard pos + 30 <= bytes.count,
              bytes[pos] == 0x50 && bytes[pos+1] == 0x4B &&
              bytes[pos+2] == 0x03 && bytes[pos+3] == 0x04 else {
            throw ZipError.invalidSignature
        }

        let fileNameLen = Int(readUInt16LE(bytes, at: pos + 26))
        let extraLen    = Int(readUInt16LE(bytes, at: pos + 28))
        let dataStart   = pos + 30 + fileNameLen + extraLen
        let compSize    = Int(entry.compressedSize)

        guard dataStart + compSize <= bytes.count else { throw ZipError.truncated }

        let compressedData = Data(bytes[dataStart..<(dataStart + compSize)])

        switch entry.compressionMethod {
        case 0:  return compressedData   // Stored
        case 8:  return try inflateDeflate(compressedData,
                                           expectedSize: Int(entry.uncompressedSize))
        default: throw ZipError.unsupportedCompression(entry.compressionMethod)
        }
    }

    // MARK: - zlib raw-deflate decompression

    private static func inflateDeflate(_ compressed: Data, expectedSize: Int) throws -> Data {
        var outSize = max(expectedSize, compressed.count * 4, 4096)
        var result = Data()
        var ok = false

        while !ok && outSize <= 64 * 1024 * 1024 {
            var buf = [UInt8](repeating: 0, count: outSize)
            var produced = 0
            var status: Int32 = Z_BUF_ERROR

            compressed.withUnsafeBytes { srcRaw in
                guard let srcPtr = srcRaw.bindMemory(to: UInt8.self).baseAddress else { return }
                var stream = z_stream()
                stream.next_in  = UnsafeMutablePointer(mutating: srcPtr)
                stream.avail_in = uInt(compressed.count)
                stream.next_out = &buf
                stream.avail_out = uInt(outSize)

                guard inflateInit2_(&stream, -MAX_WBITS, ZLIB_VERSION,
                                    Int32(MemoryLayout<z_stream>.size)) == Z_OK else { return }
                status = inflate(&stream, Z_FINISH)
                produced = outSize - Int(stream.avail_out)
                inflateEnd(&stream)
            }

            if status == Z_STREAM_END {
                result = Data(buf.prefix(produced))
                ok = true
            } else {
                outSize *= 2
            }
        }

        guard ok else { throw ZipError.decompressFailed }
        return result
    }

    // MARK: - Little-endian helpers

    private static func readUInt16LE(_ bytes: [UInt8], at offset: Int) -> UInt16 {
        guard offset + 1 < bytes.count else { return 0 }
        return UInt16(bytes[offset]) | (UInt16(bytes[offset + 1]) << 8)
    }

    private static func readUInt32LE(_ bytes: [UInt8], at offset: Int) -> UInt32 {
        guard offset + 3 < bytes.count else { return 0 }
        return UInt32(bytes[offset])
             | (UInt32(bytes[offset + 1]) << 8)
             | (UInt32(bytes[offset + 2]) << 16)
             | (UInt32(bytes[offset + 3]) << 24)
    }
}
