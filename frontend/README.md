# GOALKeeper Plan Frontend

A simple, clean frontend for the GOALKeeper Plan Notion-like workspace backend.

## Features

- **Workspace Management**: Create, view, and delete workspaces
- **Page Management**: Create, edit, and delete pages within workspaces
- **Block Editor**: Rich content editing with different block types (paragraph, heading, checklist)
- **Real-time Updates**: Changes are saved automatically
- **Clean UI**: Simple, intuitive interface

## Getting Started

### Prerequisites

- Node.js 18+ 
- npm or yarn
- Backend API running (default: http://localhost:8080)

### Installation

```bash
# Install dependencies
npm install

# Set up environment variables
# Create .env.local file with:
NEXT_PUBLIC_API_URL=http://localhost:8000/api/v1
```

### Development

```bash
# Start development server
npm run dev
```

The app will be available at http://localhost:3000

### Build

```bash
# Build for production
npm run build

# Start production server
npm start
```

## Usage

1. **Create a Workspace**: Click "New Workspace" on the home page
2. **Open Workspace**: Click "Open" on any workspace card
3. **Create Pages**: Click "+ New Page" in the sidebar
4. **Edit Pages**: Click on a page to view and edit it
5. **Add Blocks**: Type in the text area at the bottom and click "Add Block"
6. **Edit Blocks**: Click on any block to edit it
7. **Delete**: Use the delete buttons to remove workspaces, pages, or blocks

## API Integration

The frontend uses the following API endpoints:

- `GET /api/v1/notion/workspaces` - List workspaces
- `POST /api/v1/notion/workspaces` - Create workspace
- `GET /api/v1/notion/pages?workspaceId=...` - List pages
- `POST /api/v1/notion/pages` - Create page
- `GET /api/v1/notion/blocks?pageId=...` - List blocks
- `POST /api/v1/notion/blocks` - Create block
- `PUT /api/v1/notion/blocks/:id` - Update block
- `DELETE /api/v1/notion/blocks/:id` - Delete block

## Authentication

Currently, the frontend expects a `X-User-ID` header for API requests. In production, this should be handled by proper authentication middleware.

To test locally, you can:
1. Set up a test user in the backend
2. Modify the API client to include the user ID in headers
3. Or use the auth system if already configured

## Project Structure

```
src/
├── app/              # Next.js app router pages
├── components/       # React components
│   ├── BlockEditor.tsx    # Block editing component
│   ├── BlockList.tsx      # List of blocks
│   └── ui/           # UI components (buttons, cards, etc.)
├── services/         # API client services
│   ├── workspaceApi.ts   # Workspace API client
│   └── blockApi.ts       # Block API client
└── utils/            # Utility functions
```

## Technologies

- **Next.js 16** - React framework
- **TypeScript** - Type safety
- **Tailwind CSS** - Styling
- **React Query** - Data fetching (if configured)
- **Sonner** - Toast notifications

## Notes

- The frontend is designed to be simple and functional
- Error handling is basic but functional
- The UI is responsive and works on mobile devices
- Accessibility features are included in the BlockEditor component
