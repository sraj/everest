import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import type { PayloadAction } from '@reduxjs/toolkit'

export interface Document {
  id: string
  title: string
  content: string
  thumbnail_url?: string
  created_at: string
  updated_at: string
}

interface DocumentState {
  documents: Document[]
  currentDocument: Document | null
  loading: boolean
  error: string | null
}

const initialState: DocumentState = {
  documents: [],
  currentDocument: null,
  loading: false,
  error: null,
}

export const fetchDocuments = createAsyncThunk(
  'documents/fetchDocuments',
  async () => {
    const response = await fetch('/api/v1/documents')
    if (!response.ok) {
      throw new Error('Failed to fetch documents')
    }
    const data = await response.json()
    return data.documents as Document[]
  }
)

export const fetchDocument = createAsyncThunk(
  'documents/fetchDocument',
  async (id: string) => {
    const response = await fetch(`/api/v1/documents/${id}`)
    const data = await response.json()
    return data as Document
  }
)

export const deleteDocument = createAsyncThunk(
  'documents/deleteDocument',
  async (id: string) => {
    const response = await fetch(`/api/v1/documents/${id}`, {
      method: 'DELETE',
    })
    if (!response.ok) {
      throw new Error('Failed to delete document')
    }
    return id
  }
)

const documentSlice = createSlice({
  name: 'documents',
  initialState,
  reducers: {
    setCurrentDocument: (state, action: PayloadAction<Document | null>) => {
      state.currentDocument = action.payload
    },
    clearError: (state) => {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchDocuments.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchDocuments.fulfilled, (state, action) => {
        state.loading = false
        state.documents = action.payload
      })
      .addCase(fetchDocuments.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message || 'Failed to fetch documents'
      })
      .addCase(fetchDocument.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchDocument.fulfilled, (state, action) => {
        state.loading = false
        state.currentDocument = action.payload
      })
      .addCase(fetchDocument.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message || 'Failed to fetch document'
      })
      .addCase(deleteDocument.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(deleteDocument.fulfilled, (state, action) => {
        state.loading = false
        state.documents = state.documents.filter(doc => doc.id !== action.payload)
      })
      .addCase(deleteDocument.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message || 'Failed to delete document'
      })
  },
})

export const { setCurrentDocument, clearError } = documentSlice.actions
export default documentSlice.reducer
