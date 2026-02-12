import 'dart:io';
import 'package:flutter/services.dart' show rootBundle;
import 'package:dio/dio.dart';
import 'package:path_provider/path_provider.dart';
import 'package:sqflite/sqflite.dart';
import 'package:path/path.dart';
import 'package:crypto/crypto.dart';
import 'dart:convert';

/// Manages the local SQLite database, including downloading updates and querying.
class DatabaseManager {
  static final DatabaseManager _instance = DatabaseManager._internal();
  factory DatabaseManager() => _instance;
  DatabaseManager._internal();

  Database? _db;
  final Dio _dio = Dio();
  
  // URL to the raw files hosted on GitHub Pages or raw.githubusercontent.com
  // Replace with your actual repo details
  final String _repoBaseUrl = 'https://raw.githubusercontent.com/entreya/job-aggregation/main';
  
  /// Ensures the database is initialized and ready.
  Future<Database> get database async {
    if (_db != null && _db!.isOpen) return _db!;
    _db = await _initDatabase();
    return _db!;
  }

  /// Initializes the database.
  Future<Database> _initDatabase() async {
    final dbPath = await _getDbPath();
    
    // Copy from assets if not exists
    if (!await File(dbPath).exists()) {
      print("Copying seeded jobs.db from assets...");
      try {
        // Load from assets
        // Note: In real app, use rootBundle.load. 
        // For CLI/Dart IO it's different, but this is execution on device.
        // We need 'package:flutter/services.dart'
        
        final byteData = await rootBundle.load('assets/jobs.db');
        final bytes = byteData.buffer.asUint8List(byteData.offsetInBytes, byteData.lengthInBytes);
        await File(dbPath).writeAsBytes(bytes);
        print("Seeded database copied.");
      } catch (e) {
        print("Failed to copy asset DB: $e");
      }
    }
    
    return await openDatabase(dbPath, version: 1);
  }

  Future<String> _getDbPath() async {
    final docsDir = await getApplicationDocumentsDirectory();
    return join(docsDir.path, 'jobs.db');
  }

  /// Checks for updates and downloads the new DB if available.
  /// Returns true if an update was applied.
  Future<bool> checkForUpdates() async {
    try {
      // 1. Fetch Metadata
      final response = await _dio.get('$_repoBaseUrl/metadata.json');
      final serverMeta = response.data; // automatic json decoding by Dio
      
      final String serverChecksum = serverMeta['checksum'];
      final int serverTimestamp = serverMeta['last_updated'];

      // 2. Check Local Metadata (if exists)
      final docsDir = await getApplicationDocumentsDirectory();
      final metaFile = File(join(docsDir.path, 'metadata.json'));
      
      if (await metaFile.exists()) {
        final localMetaJson = await metaFile.readAsString();
        final localMeta = jsonDecode(localMetaJson);
        if (localMeta['checksum'] == serverChecksum) {
          print("Database is up to date.");
          return false; 
        }
      }

      print("New database version found ($serverTimestamp). Downloading...");
      await _downloadAndReplaceDB(serverChecksum);
      
      // Update local metadata
      await metaFile.writeAsString(jsonEncode(serverMeta));
      
      return true;

    } catch (e) {
      print("Failed to check for updates: $e");
      return false;
    }
  }

  /// Downloads the DB to a temp file, verifies checksum, and hot-swaps.
  Future<void> _downloadAndReplaceDB(String expectedChecksum) async {
    final docsDir = await getApplicationDocumentsDirectory();
    final tempDbPath = join(docsDir.path, 'temp_jobs.db');
    final finalDbPath = join(docsDir.path, 'jobs.db');

    // 1. Download to temp file
    await _dio.download(
      '$_repoBaseUrl/jobs.db',
      tempDbPath,
      onReceiveProgress: (received, total) {
        if (total != -1) {
          // print((received / total * 100).toStringAsFixed(0) + "%");
        }
      },
    );

    // 2. Verify Checksum
    final file = File(tempDbPath);
    final bytes = await file.readAsBytes();
    final digest = sha256.convert(bytes);
    
    if (digest.toString() != expectedChecksum) {
      await file.delete();
      throw Exception("Checksum mismatch! Downloaded DB might be corrupted.");
    }

    // 3. Hot-Swap
    // Close existing connection if open
    if (_db != null && _db!.isOpen) {
      await _db!.close();
      _db = null;
    }

    // Replace file
    // Windows might lock file, but on Mobile (iOS/Android) this is usually fine if closed.
    await file.rename(finalDbPath);
    print("Database updated successfully.");
    
    // Re-initialize connection
    await _initDatabase();
  }

  /// Search jobs by title or department.
  Future<List<Map<String, dynamic>>> searchJobs(String query) async {
    final db = await database;
    final sanitizedQuery = '%$query%';
    
    // Ensure table exists (in case we opened a fresh empty DB before first download)
    // A robust app would have the schema created in _initDatabase via onCreate 
    // or loaded from asset. Assuming downloaded DB has schema.
    
    try {
        final results = await db.rawQuery('''
          SELECT * FROM jobs 
          WHERE title LIKE ? OR department LIKE ? 
          ORDER BY posted_date DESC
        ''', [sanitizedQuery, sanitizedQuery]);
        
        return results;
    } catch (e) {
        // Table might not exist yet if download hasn't finished
        print("Query failed: $e");
        return [];
    }
  }
}
