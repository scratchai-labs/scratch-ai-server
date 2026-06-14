import path from 'node:path'

const SAMPLE_PROJECT_JSON = `{
  "targets": [
    {
      "isStage": true,
      "name": "Stage",
      "broadcasts": {
        "message1": "开始"
      },
      "variables": {
        "score": ["分数", 0]
      },
      "lists": {
        "todo": ["步骤列表", []]
      },
      "blocks": {
        "stage-1": { "opcode": "event_whenflagclicked", "next": "stage-2", "topLevel": true },
        "stage-2": { "opcode": "event_broadcast", "next": null, "parent": "stage-1" }
      }
    },
    {
      "isStage": false,
      "name": "Cat",
      "blocks": {
        "cat-1": { "opcode": "event_whenbroadcastreceived", "next": "cat-2", "topLevel": true },
        "cat-2": { "opcode": "motion_movesteps", "next": "cat-3", "parent": "cat-1" },
        "cat-3": { "opcode": "motion_ifonedgebounce", "next": null, "parent": "cat-2" },
        "cat-4": { "opcode": "looks_say", "next": null, "topLevel": true }
      }
    }
  ],
  "extensions": ["pen"]
}`

const CRC32_TABLE = buildCrc32Table()

export function buildRealSmokeApiEnv({
  inheritedEnv,
  apiPort,
  webOrigin,
  tempDir,
}) {
  return {
    ...inheritedEnv,
    PORT: String(apiPort),
    GIN_MODE: 'debug',
    DATABASE_URL: '',
    SERVER_API_DB_PATH: path.join(tempDir, 'server-api.sqlite3'),
    SB3_STORAGE_DIR: path.join(tempDir, 'sb3-storage'),
    CORS_ALLOWED_ORIGINS: webOrigin,
    ADMIN_BOOTSTRAP_USERNAME: '',
    ADMIN_BOOTSTRAP_PASSWORD: '',
  }
}

export function buildRealSmokeWebEnv({
  inheritedEnv,
  apiBaseUrl,
}) {
  return {
    ...inheritedEnv,
    VITE_SERVER_WEB_API_MODE: 'real',
    VITE_SERVER_WEB_API_BASE_URL: apiBaseUrl,
  }
}

export function createSampleSb3Archive() {
  const buffer = createStoredZip([
    {
      name: 'project.json',
      data: Buffer.from(SAMPLE_PROJECT_JSON, 'utf8'),
    },
  ])

  return {
    fileName: 'teacher-real-smoke.sb3',
    contentType: 'application/zip',
    buffer,
  }
}

function createStoredZip(entries) {
  const localParts = []
  const centralParts = []
  let localOffset = 0

  for (const entry of entries) {
    const nameBuffer = Buffer.from(entry.name, 'utf8')
    const dataBuffer = Buffer.from(entry.data)
    const crc = crc32(dataBuffer)

    const localHeader = Buffer.alloc(30)
    localHeader.writeUInt32LE(0x04034b50, 0)
    localHeader.writeUInt16LE(20, 4)
    localHeader.writeUInt16LE(0, 6)
    localHeader.writeUInt16LE(0, 8)
    localHeader.writeUInt16LE(0, 10)
    localHeader.writeUInt16LE(0, 12)
    localHeader.writeUInt32LE(crc, 14)
    localHeader.writeUInt32LE(dataBuffer.length, 18)
    localHeader.writeUInt32LE(dataBuffer.length, 22)
    localHeader.writeUInt16LE(nameBuffer.length, 26)
    localHeader.writeUInt16LE(0, 28)
    localParts.push(localHeader, nameBuffer, dataBuffer)

    const centralHeader = Buffer.alloc(46)
    centralHeader.writeUInt32LE(0x02014b50, 0)
    centralHeader.writeUInt16LE(20, 4)
    centralHeader.writeUInt16LE(20, 6)
    centralHeader.writeUInt16LE(0, 8)
    centralHeader.writeUInt16LE(0, 10)
    centralHeader.writeUInt16LE(0, 12)
    centralHeader.writeUInt16LE(0, 14)
    centralHeader.writeUInt32LE(crc, 16)
    centralHeader.writeUInt32LE(dataBuffer.length, 20)
    centralHeader.writeUInt32LE(dataBuffer.length, 24)
    centralHeader.writeUInt16LE(nameBuffer.length, 28)
    centralHeader.writeUInt16LE(0, 30)
    centralHeader.writeUInt16LE(0, 32)
    centralHeader.writeUInt16LE(0, 34)
    centralHeader.writeUInt16LE(0, 36)
    centralHeader.writeUInt32LE(0, 38)
    centralHeader.writeUInt32LE(localOffset, 42)
    centralParts.push(centralHeader, nameBuffer)

    localOffset += localHeader.length + nameBuffer.length + dataBuffer.length
  }

  const centralDirectory = Buffer.concat(centralParts)
  const endRecord = Buffer.alloc(22)
  endRecord.writeUInt32LE(0x06054b50, 0)
  endRecord.writeUInt16LE(0, 4)
  endRecord.writeUInt16LE(0, 6)
  endRecord.writeUInt16LE(entries.length, 8)
  endRecord.writeUInt16LE(entries.length, 10)
  endRecord.writeUInt32LE(centralDirectory.length, 12)
  endRecord.writeUInt32LE(localOffset, 16)
  endRecord.writeUInt16LE(0, 20)

  return Buffer.concat([...localParts, centralDirectory, endRecord])
}

function crc32(buffer) {
  let value = 0xffffffff

  for (const byte of buffer) {
    value = (value >>> 8) ^ CRC32_TABLE[(value ^ byte) & 0xff]
  }

  return (value ^ 0xffffffff) >>> 0
}

function buildCrc32Table() {
  const table = new Uint32Array(256)

  for (let index = 0; index < table.length; index += 1) {
    let value = index
    for (let bit = 0; bit < 8; bit += 1) {
      value = (value & 1) === 1 ? (value >>> 1) ^ 0xedb88320 : value >>> 1
    }
    table[index] = value >>> 0
  }

  return table
}
